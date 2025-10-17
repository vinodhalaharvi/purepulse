// cmd/test-team-weekly/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/vinodhalaharvi/purekernels/pkg/effect"
	"github.com/vinodhalaharvi/purekernels/pkg/result"
	"github.com/vinodhalaharvi/purepulse/db"
	"github.com/vinodhalaharvi/purepulse/db/query"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
	"github.com/vinodhalaharvi/purepulse/pkg/llm"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// TYPE ALIASES
// ============================================================================

type BeginTime time.Time
type EndTime time.Time
type UserCount int
type SuccessCount int

// ============================================================================
// USER ANALYSIS TASK AS EFFECT
// ============================================================================

// UserAnalysisResult wraps a single user's analysis
type UserAnalysisResult struct {
	UserID types.UserID
	Report analytics.UserWeeklyReport
	Error  error
}

// UserAnalysisResultMonoid combines multiple user analysis results
type UserAnalysisResultMonoid struct{}

func (UserAnalysisResultMonoid) Empty() UserAnalysisResult {
	return UserAnalysisResult{}
}

func (UserAnalysisResultMonoid) Combine(a, b UserAnalysisResult) UserAnalysisResult {
	if a.Error != nil {
		return b
	}
	return a
}

// ============================================================================
// BUILD USER ANALYSIS TASK USING QUERY MONOIDS
// ============================================================================

func BuildUserAnalysisEffect(
	ctx context.Context,
	conn *db.Connection,
	claudeClient *llm.ClaudeClient,
	userID types.UserID,
	beginTime BeginTime,
	endTime EndTime,
) effect.Writer[[]string, result.Result[UserAnalysisResult]] {

	logs := []string{fmt.Sprintf("build_analysis_started: user=%s", userID)}

	// Build query using monoids
	userFilter := query.Where("user_id", string(userID))
	timeFilter := query.WhereBetween("date", time.Time(beginTime), time.Time(endTime))
	combinedFilter := userFilter.Combine(timeFilter)

	selectFields := query.Select(
		"user_id", "date", "source", "type", "event_count", "first_event_at", "last_event_at",
		"channels", "collaborators", "total_size", "total_duration_seconds",
	)

	dailyActivityBuilder := query.Table[query.DailyActivityAgg]("daily_user_activity").
		Project(selectFields).
		Filter(combinedFilter).
		Sort(query.Desc("date")).
		Bound(query.Limit(50))

	sql, params := dailyActivityBuilder.Build()
	logs = append(logs, fmt.Sprintf("query_built: %s", sql))

	// Execute query
	userRows, err := conn.DB.QueryContext(ctx, sql, params...)
	if err != nil {
		logs = append(logs, fmt.Sprintf("query_failed: %v", err))
		return effect.NewWriter(
			result.Err[UserAnalysisResult](err),
			logs,
		)
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
		logs = append(logs, "no_activity_found")
		return effect.NewWriter(
			result.Err[UserAnalysisResult](fmt.Errorf("no activity for %s", userID)),
			logs,
		)
	}

	logs = append(logs, fmt.Sprintf("fetched_activity: rows=%d", len(dailyActivity)))

	// Analyze with Claude
	week := types.TimeRange{
		Start: time.Time(beginTime),
		End:   time.Time(endTime),
	}

	analysisWriter := llm.AnalyzeWeeklyActivity(ctx, claudeClient, userID, week, dailyActivity)
	analysisRes, analysisLogs := analysisWriter.Run()
	logs = append(logs, analysisLogs...)

	if !analysisRes.IsOk() {
		logs = append(logs, fmt.Sprintf("analysis_failed: %v", analysisRes.Error()))
		return effect.NewWriter(
			result.Err[UserAnalysisResult](analysisRes.Error()),
			logs,
		)
	}

	report := analysisRes.Unwrap()
	analysisResult := UserAnalysisResult{
		UserID: userID,
		Report: report,
		Error:  nil,
	}

	logs = append(logs, fmt.Sprintf("analysis_succeeded: wins=%d, blocked=%d", len(report.Wins), len(report.Blocked)))
	return effect.NewWriter(result.Ok(analysisResult), logs)
}

// ============================================================================
// TRAVERSE FOR PARALLEL EXECUTION
// ============================================================================

// TraverseUserAnalysesParallel executes multiple user analyses in parallel
func TraverseUserAnalysesParallel(
	ctx context.Context,
	conn *db.Connection,
	claudeClient *llm.ClaudeClient,
	userIDs []types.UserID,
	beginTime BeginTime,
	endTime EndTime,
) effect.Writer[[]string, result.Result[map[types.UserID]analytics.UserWeeklyReport]] {

	logs := []string{fmt.Sprintf("traverse_started: users=%d", len(userIDs))}

	// Create channels for parallel execution
	type taskResult struct {
		userID types.UserID
		res    result.Result[UserAnalysisResult]
		logs   []string
	}

	resultsChan := make(chan taskResult, len(userIDs))

	// Execute all tasks in parallel (goroutines)
	for _, userID := range userIDs {
		go func(uid types.UserID) {
			taskWriter := BuildUserAnalysisEffect(ctx, conn, claudeClient, uid, beginTime, endTime)
			res, taskLogs := taskWriter.Run()
			resultsChan <- taskResult{
				userID: uid,
				res:    res,
				logs:   taskLogs,
			}
		}(userID)
	}

	// Collect results using monoid
	userReports := make(map[types.UserID]analytics.UserWeeklyReport)
	successCount := SuccessCount(0)

	for i := 0; i < len(userIDs); i++ {
		tr := <-resultsChan
		logs = append(logs, tr.logs...)

		if tr.res.IsOk() {
			analysisResult := tr.res.Unwrap()
			if analysisResult.Error == nil {
				userReports[tr.userID] = analysisResult.Report
				successCount++
			}
		}
	}

	logs = append(logs, fmt.Sprintf(
		"traverse_completed: succeeded=%d, failed=%d",
		successCount,
		len(userIDs)-int(successCount),
	))

	if len(userReports) == 0 {
		return effect.NewWriter(
			result.Err[map[types.UserID]analytics.UserWeeklyReport](fmt.Errorf("all analyses failed")),
			logs,
		)
	}

	return effect.NewWriter(result.Ok(userReports), logs)
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	ctx := context.Background()

	fmt.Println("Team Weekly Aggregation (Traverse + Query Monoids)")
	fmt.Println("==================================================\n")

	// Connect
	fmt.Println("Step 0: Connecting to database...")
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
	// STEP 1: Fetch team members
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
	// STEP 2: Execute all user analyses in parallel using Traverse
	// ========================================================================
	fmt.Println("Step 2: Executing user analyses in parallel...")

	claudeClient := &llm.ClaudeClient{
		APIKey:     os.Getenv("ANTHROPIC_API_KEY"),
		Model:      "claude-sonnet-4-20250514",
		BaseURL:    "https://api.anthropic.com/v1",
		MaxRetries: 3,
		Timeout:    60 * time.Second,
	}

	beginTime := BeginTime(time.Now().AddDate(0, 0, -7))
	endTime := EndTime(time.Now())

	traverseWriter := TraverseUserAnalysesParallel(ctx, conn, claudeClient, teamMembers, beginTime, endTime)
	traverseRes, traverseLogs := traverseWriter.Run()

	fmt.Println("Traverse logs:")
	for _, log := range traverseLogs {
		fmt.Printf("  %s\n", log)
	}

	if !traverseRes.IsOk() {
		log.Fatalf("Traverse failed: %v", traverseRes.Error())
	}

	userReports := traverseRes.Unwrap()
	fmt.Printf("\nGenerated %d user reports\n\n", len(userReports))

	// ========================================================================
	// STEP 3: Aggregate to team metrics
	// ========================================================================
	fmt.Println("Step 3: Aggregating to team metrics...")

	teamID := types.TeamID("team-engineering")
	week := types.TimeRange{
		Start: time.Time(beginTime),
		End:   time.Time(endTime),
	}

	aggregateWriter := analytics.FetchAndAggregateWeeklyTeamReports(ctx, teamID, week, userReports)
	aggregateRes, aggregateLogs := aggregateWriter.Run()

	for _, log := range aggregateLogs {
		fmt.Printf("  %s\n", log)
	}

	if !aggregateRes.IsOk() {
		log.Fatalf("Aggregation failed: %v", aggregateRes.Error())
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
	fmt.Println("Step 4: Generating team-level Claude analysis...")

	teamAnalysisWriter := llm.AnalyzeTeamWeekly(ctx, claudeClient, teamID, week, metrics, userReports)
	teamAnalysisRes, teamAnalysisLogs := teamAnalysisWriter.Run()

	fmt.Println("Team analysis logs:")
	for _, log := range teamAnalysisLogs {
		fmt.Printf("  %s\n", log)
	}

	if !teamAnalysisRes.IsOk() {
		log.Fatalf("Team analysis failed: %v", teamAnalysisRes.Error())
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
