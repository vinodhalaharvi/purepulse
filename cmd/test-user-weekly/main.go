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
	"github.com/vinodhalaharvi/purepulse/pkg/llm"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	ctx := context.Background()

	fmt.Println("User Weekly Reports Test")
	fmt.Println("=======================\n")

	// Connect to DB
	dbConfig, err := db.NewConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	conn, err := db.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connected to database\n")

	// Setup Claude
	claudeClient := &llm.ClaudeClient{
		APIKey:     os.Getenv("ANTHROPIC_API_KEY"),
		Model:      "claude-sonnet-4-20250514",
		BaseURL:    "https://api.anthropic.com/v1",
		MaxRetries: 3,
		Timeout:    60 * time.Second,
	}

	// Fetch all users
	fmt.Println("Fetching team members...")
	userQuery := `SELECT DISTINCT user_id FROM events`
	rows, err := conn.DB.QueryContext(ctx, userQuery)
	if err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}
	defer rows.Close()

	var userIDs []types.UserID
	for rows.Next() {
		var userID types.UserID
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		userIDs = append(userIDs, userID)
	}

	fmt.Printf("Found %d users\n\n", len(userIDs))

	week := types.TimeRange{
		Start: time.Now().AddDate(0, 0, -7),
		End:   time.Now(),
	}

	fmt.Printf("Analyzing week: %s to %s\n\n", week.Start.Format("2006-01-02"), week.End.Format("2006-01-02"))

	// Process each user
	for i, userID := range userIDs {
		fmt.Printf("[%d/%d] Analyzing user: %s\n", i+1, len(userIDs), userID)

		// Fetch daily activity
		dailyActivityQuery := `
            SELECT user_id, date, source, type, event_count, first_event_at, last_event_at,
                   channels, collaborators, total_size, total_duration_seconds
            FROM daily_user_activity
            WHERE user_id = $1 AND date >= $2 AND date <= $3
            ORDER BY date DESC
            LIMIT 50
        `

		activityRows, err := conn.DB.QueryContext(ctx, dailyActivityQuery, userID, week.Start, week.End)
		if err != nil {
			fmt.Printf("  ✗ Query failed: %v\n", err)
			continue
		}

		var dailyActivity []query.DailyActivityAgg
		for activityRows.Next() {
			row, err := query.ScanDailyActivityAgg(activityRows)
			if err != nil {
				continue
			}
			dailyActivity = append(dailyActivity, row)
		}
		activityRows.Close()

		if len(dailyActivity) == 0 {
			fmt.Printf("  ✗ No activity found\n")
			continue
		}

		fmt.Printf("  ✓ Found %d activity records\n", len(dailyActivity))

		// Analyze with Claude
		analysisWriter := llm.AnalyzeWeeklyActivity(ctx, claudeClient, userID, week, dailyActivity)
		analysisRes, _ := analysisWriter.Run()

		if !analysisRes.IsOk() {
			fmt.Printf("  ✗ Analysis failed: %v\n", analysisRes.Error())
			continue
		}

		report := analysisRes.Unwrap()

		// Save to database
		saveWriter := llm.SaveUserWeeklyReport(ctx, conn, userID, week, report, claudeClient.Model, 0, 0)
		saveRes, saveLogs := saveWriter.Run()

		fmt.Println("  Save logs:")
		for _, log := range saveLogs {
			fmt.Printf("    %s\n", log)
		}

		if !saveRes.IsOk() {
			fmt.Printf("  ✗ Save failed: %v\n", saveRes.Error())
			continue
		}

		fmt.Printf("  ✓ Wins: %d, Blocked: %d, In Progress: %d\n", len(report.Wins), len(report.Blocked), len(report.InProgress))

		// Display individual report
		fmt.Printf("\n  📝 SUMMARY:\n  %s\n", report.Notes)

		if len(report.Wins) > 0 {
			fmt.Printf("\n  ✅ WINS:\n")
			for j, w := range report.Wins {
				fmt.Printf("    %d. %s\n", j+1, w.Title)
				fmt.Printf("       %s\n", w.Impact)
			}
		}

		if len(report.InProgress) > 0 {
			fmt.Printf("\n  🔄 IN PROGRESS:\n")
			for j, p := range report.InProgress {
				fmt.Printf("    %d. %s (%d%%)\n", j+1, p.Title, p.PercentComplete)
			}
		}

		if len(report.Blocked) > 0 {
			fmt.Printf("\n  🚫 BLOCKED:\n")
			for j, b := range report.Blocked {
				fmt.Printf("    %d. %s\n", j+1, b.Title)
				fmt.Printf("       Action: %s\n", b.SuggestedAction)
			}
		}

		fmt.Printf("\n")
	}

	fmt.Printf("\n✨ All user reports processed!\n")

	// Verify all saved
	verifyQuery := `SELECT COUNT(*) FROM weekly_reports WHERE week_start = $1`
	var count int
	conn.DB.QueryRowContext(ctx, verifyQuery, week.Start).Scan(&count)
	fmt.Printf("Total reports saved: %d\n", count)
}
