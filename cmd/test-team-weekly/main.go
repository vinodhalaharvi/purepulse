package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/vinodhalaharvi/purepulse/db"
	"github.com/vinodhalaharvi/purepulse/db/query"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
	"github.com/vinodhalaharvi/purepulse/pkg/llm"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	ctx := context.Background()

	fmt.Println("Team Weekly Aggregation Test")
	fmt.Println("============================\n")

	// Connect
	fmt.Println("Connecting to database...")
	dbConfig, err := db.NewConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	conn, err := db.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connected\n")

	// ========================================================================
	// STEP 1: Fetch all users
	// ========================================================================
	fmt.Println("Step 1: Fetching team members...")

	userQuery := `SELECT DISTINCT user_id FROM events LIMIT 10`
	rows, err := conn.DB.QueryContext(ctx, userQuery)
	if err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}
	defer rows.Close()

	var teamMembers []types.UserID
	for rows.Next() {
		var userID types.UserID
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		teamMembers = append(teamMembers, userID)
	}

	fmt.Printf("Found %d team members: %v\n\n", len(teamMembers), teamMembers)

	// ========================================================================
	// STEP 2: Fetch individual user reports using Claude
	// ========================================================================
	fmt.Println("Step 2: Generating individual user reports...")

	claudeClient := &llm.ClaudeClient{
		APIKey:     os.Getenv("ANTHROPIC_API_KEY"),
		Model:      "claude-sonnet-4-20250514",
		BaseURL:    "https://api.anthropic.com/v1",
		MaxRetries: 3,
		Timeout:    60 * time.Second,
	}

	userReports := make(map[types.UserID]analytics.UserWeeklyReport)
	week := types.TimeRange{
		Start: time.Now().AddDate(0, 0, -7),
		End:   time.Now(),
	}

	// Fetch daily activity for each user
	dailyActivityQuery := `
        SELECT user_id, date, source, type, event_count, first_event_at, last_event_at, 
               channels, collaborators, total_size, total_duration_seconds
        FROM daily_user_activity
        WHERE user_id = $1 AND date >= $2 AND date <= $3
        ORDER BY date DESC
        LIMIT 50
    `

	for _, userID := range teamMembers {
		fmt.Printf("\n  Analyzing user: %s", userID)

		userRows, err := conn.DB.QueryContext(ctx, dailyActivityQuery, userID, week.Start, week.End)
		if err != nil {
			fmt.Printf(" - Query failed: %v\n", err)
			continue
		}

		var dailyActivity []query.DailyActivityAgg
		for userRows.Next() {
			row, err := query.ScanDailyActivityAgg(userRows)
			if err != nil {
				continue
			}
			dailyActivity = append(dailyActivity, row)
		}
		userRows.Close()

		if len(dailyActivity) == 0 {
			fmt.Println(" - No activity")
			continue
		}

		// Analyze with Claude
		analysisWriter := llm.AnalyzeWeeklyActivity(ctx, claudeClient, userID, week, dailyActivity)
		analysisRes, _ := analysisWriter.Run()

		if !analysisRes.IsOk() {
			fmt.Printf(" - Claude analysis failed: %v\n", analysisRes.Error())
			continue
		}

		report := analysisRes.Unwrap()
		userReports[userID] = report
		fmt.Printf(" - Success (%d wins, %d blocked)\n", len(report.Wins), len(report.Blocked))
	}

	fmt.Printf("\nGenerated %d user reports\n\n", len(userReports))

	if len(userReports) == 0 {
		log.Fatal("No user reports generated")
	}

	// ========================================================================
	// STEP 3: Aggregate user reports to team metrics
	// ========================================================================
	fmt.Println("Step 3: Aggregating to team metrics...")

	teamID := types.TeamID("team-engineering")

	aggregateWriter := analytics.FetchAndAggregateWeeklyTeamReports(ctx, teamID, week, userReports)
	aggregateRes, aggregateLogs := aggregateWriter.Run()

	for _, log := range aggregateLogs {
		fmt.Printf("  %s\n", log)
	}

	if !aggregateRes.IsOk() {
		log.Fatalf("Failed to aggregate: %v", aggregateRes.Error())
	}

	metrics := aggregateRes.Unwrap()
	fmt.Printf("\nTeam Metrics:\n")
	fmt.Printf("  Users: %d\n", metrics.TotalUserCount)
	fmt.Printf("  Total Wins: %d\n", metrics.TotalWinsReported)
	fmt.Printf("  Total Blockers: %d\n", metrics.TotalBlockersReported)
	fmt.Printf("  Avg Productivity: %.2f\n", metrics.AverageProductivityScore)
	fmt.Printf("  Avg Collaboration: %.2f\n\n", metrics.AverageCollaborationScore)

	// ========================================================================
	// STEP 4: Generate team-level Claude analysis
	// ========================================================================
	fmt.Println("Step 4: Generating team-level analysis...")

	teamAnalysisWriter := llm.AnalyzeTeamWeekly(ctx, claudeClient, teamID, week, metrics, userReports)
	teamAnalysisRes, teamAnalysisLogs := teamAnalysisWriter.Run()

	fmt.Println("Team analysis logs:")
	for _, log := range teamAnalysisLogs {
		fmt.Printf("  %s\n", log)
	}

	if !teamAnalysisRes.IsOk() {
		log.Fatalf("Failed to analyze team: %v", teamAnalysisRes.Error())
	}

	teamReport := teamAnalysisRes.Unwrap()

	// ========================================================================
	// STEP 5: Display team report
	// ========================================================================
	fmt.Printf("\n════════════════════════════════════════════════════════════\n")
	fmt.Printf("TEAM WEEKLY REPORT: %s\n", teamID)
	fmt.Printf("Week: %s to %s\n", week.Start.Format("2006-01-02"), week.End.Format("2006-01-02"))
	fmt.Printf("════════════════════════════════════════════════════════════\n\n")

	fmt.Printf("📝 EXECUTIVE SUMMARY:\n%s\n\n", teamReport.ExecutiveSummary)

	fmt.Printf("📈 VELOCITY ANALYSIS:\n%s\n\n", teamReport.VelocityAnalysis)

	fmt.Printf("🤝 COLLABORATION:\n%s\n\n", teamReport.CollaborationNotes)

	if len(teamReport.TeamBlockers) > 0 {
		fmt.Printf("🚫 TEAM BLOCKERS:\n")
		for i, blocker := range teamReport.TeamBlockers {
			fmt.Printf("  %d. %s\n", i+1, blocker.Title)
			fmt.Printf("     Affected: %v\n", blocker.AffectedUsers)
			fmt.Printf("     Action: %s\n\n", blocker.Action)
		}
	}

	if len(teamReport.Recommendations) > 0 {
		fmt.Printf("💡 TEAM RECOMMENDATIONS:\n")
		for i, rec := range teamReport.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("✨ Team report generated successfully!\n")
}
