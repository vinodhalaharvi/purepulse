// pkg/analytics/test/main.go
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
	"github.com/vinodhalaharvi/purepulse/pkg/llm"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// TYPE ALIASES FOR CLARITY
// ============================================================================

type EventCount int
type PatternCount int
type ViewRowCount int

// ============================================================================
// QUERY BUILDERS USING MONOIDS
// ============================================================================

// QueryDailyUserActivity builds query using monoids
func QueryDailyUserActivity(limit int) query.Query[query.DailyActivityAgg] {
	return query.Table[query.DailyActivityAgg]("daily_user_activity").
		Sort(query.Desc("date")).
		Bound(query.Limit(limit))
}

// QueryUserCorrelationSummary builds query using monoids
// QueryUserCorrelationSummary builds query using monoids
func QueryUserCorrelationSummary(limit int) query.Query[query.CorrelationRow] {
	return query.Table[query.CorrelationRow]("user_correlation_summary").
		Sort(query.Desc("last_detected_at")). // Changed: detected_at → last_detected_at
		Bound(query.Limit(limit))
}

// QueryWeeklyTeamActivity builds query using monoids
func QueryWeeklyTeamActivity(limit int) query.Query[query.TeamActivityRow] {
	return query.Table[query.TeamActivityRow]("weekly_team_activity").
		Sort(query.Desc("week_start")).
		Bound(query.Limit(limit))
}

// RefreshView executes REFRESH MATERIALIZED VIEW
func RefreshView(
	ctx context.Context,
	conn *db.Connection,
	viewName string,
) effect.Writer[[]string, result.Result[string]] {

	logs := []string{fmt.Sprintf("refresh_view_started: view=%s", viewName)}

	// Try concurrent first
	refreshSQL := fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", viewName)
	_, err := conn.DB.ExecContext(ctx, refreshSQL)

	if err != nil {
		// Fall back to non-concurrent
		refreshSQL = fmt.Sprintf("REFRESH MATERIALIZED VIEW %s", viewName)
		_, err = conn.DB.ExecContext(ctx, refreshSQL)

		if err != nil {
			logs = append(logs, fmt.Sprintf("refresh_failed: %v", err))
			return effect.NewWriter(
				result.Err[string](fmt.Errorf("failed to refresh %s: %w", viewName, err)),
				logs,
			)
		}

		logs = append(logs, fmt.Sprintf("refresh_succeeded (non-concurrent): view=%s", viewName))
		return effect.NewWriter(result.Ok(fmt.Sprintf("Refreshed %s", viewName)), logs)
	}

	logs = append(logs, fmt.Sprintf("refresh_succeeded (concurrent): view=%s", viewName))
	return effect.NewWriter(result.Ok(fmt.Sprintf("Refreshed %s", viewName)), logs)
}

// ============================================================================
// FETCH VIEW DATA (Pure Query Building + Impure Execution)
// ============================================================================

// FetchDailyUserActivityData fetches and aggregates using monoids
func FetchDailyUserActivityData(
	ctx context.Context,
	conn *db.Connection,
	limit int,
) effect.Writer[[]string, result.Result[[]query.DailyActivityAgg]] {

	logs := []string{fmt.Sprintf("fetch_daily_activity_started: limit=%d", limit)}

	// Build query using monoids
	queryBuilder := QueryDailyUserActivity(limit)
	sql, params := queryBuilder.Build()

	logs = append(logs, fmt.Sprintf("generated_sql: %s", sql))

	rows, err := conn.DB.QueryContext(ctx, sql, params...)
	if err != nil {
		logs = append(logs, fmt.Sprintf("query_failed: %v", err))
		return effect.NewWriter(result.Err[[]query.DailyActivityAgg](err), logs)
	}
	defer rows.Close()

	var results []query.DailyActivityAgg
	for rows.Next() {
		row, err := query.ScanDailyActivityAgg(rows)
		if err != nil {
			logs = append(logs, fmt.Sprintf("scan_error: %v", err))
			continue
		}
		results = append(results, row)
	}

	logs = append(logs, fmt.Sprintf("fetch_succeeded: rows=%d", len(results)))
	return effect.NewWriter(result.Ok(results), logs)
}

// FetchCorrelationData fetches correlation summary using monoids
func FetchCorrelationData(
	ctx context.Context,
	conn *db.Connection,
	limit int,
) effect.Writer[[]string, result.Result[[]query.CorrelationRow]] {

	logs := []string{fmt.Sprintf("fetch_correlations_started: limit=%d", limit)}

	// Build query using monoids
	queryBuilder := QueryUserCorrelationSummary(limit)
	sql, params := queryBuilder.Build()

	logs = append(logs, fmt.Sprintf("generated_sql: %s", sql))

	rows, err := conn.DB.QueryContext(ctx, sql, params...)
	if err != nil {
		logs = append(logs, fmt.Sprintf("query_failed: %v", err))
		return effect.NewWriter(result.Err[[]query.CorrelationRow](err), logs)
	}
	defer rows.Close()

	var results []query.CorrelationRow
	for rows.Next() {
		row, err := query.ScanCorrelationRow(rows)
		if err != nil {
			logs = append(logs, fmt.Sprintf("scan_error: %v", err))
			continue
		}
		results = append(results, row)
	}

	logs = append(logs, fmt.Sprintf("fetch_succeeded: rows=%d", len(results)))
	return effect.NewWriter(result.Ok(results), logs)
}

// FetchTeamActivityData fetches team activity using monoids
func FetchTeamActivityData(
	ctx context.Context,
	conn *db.Connection,
	limit int,
) effect.Writer[[]string, result.Result[[]query.TeamActivityRow]] {

	logs := []string{fmt.Sprintf("fetch_team_activity_started: limit=%d", limit)}

	// Build query using monoids
	queryBuilder := QueryWeeklyTeamActivity(limit)
	sql, params := queryBuilder.Build()

	logs = append(logs, fmt.Sprintf("generated_sql: %s", sql))

	rows, err := conn.DB.QueryContext(ctx, sql, params...)
	if err != nil {
		logs = append(logs, fmt.Sprintf("query_failed: %v", err))
		return effect.NewWriter(result.Err[[]query.TeamActivityRow](err), logs)
	}
	defer rows.Close()

	var results []query.TeamActivityRow
	for rows.Next() {
		row, err := query.ScanTeamActivityRow(rows)
		if err != nil {
			logs = append(logs, fmt.Sprintf("scan_error: %v", err))
			continue
		}
		results = append(results, row)
	}

	logs = append(logs, fmt.Sprintf("fetch_succeeded: rows=%d", len(results)))
	return effect.NewWriter(result.Ok(results), logs)
}

// CheckViewData counts rows in a view using monoids
func CheckViewData(
	ctx context.Context,
	conn *db.Connection,
	viewName string,
) effect.Writer[[]string, result.Result[ViewRowCount]] {

	logs := []string{fmt.Sprintf("check_view_started: view=%s", viewName)}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", viewName)
	var count int

	err := conn.DB.QueryRowContext(ctx, countQuery).Scan(&count)
	if err != nil {
		logs = append(logs, fmt.Sprintf("check_failed: %v", err))
		return effect.NewWriter(result.Err[ViewRowCount](err), logs)
	}

	logs = append(logs, fmt.Sprintf("check_succeeded: row_count=%d", count))
	return effect.NewWriter(result.Ok(ViewRowCount(count)), logs)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	ctx := context.Background()

	fmt.Println("Starting Weekly Metrics Diagnostic")
	fmt.Println("==================================\n")

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
	// STEP 1: Check base events
	// ========================================================================
	fmt.Println("Step 1: Checking raw events table...")

	eventCountQuery := "SELECT COUNT(*) FROM events"
	var eventCount int
	if err := conn.DB.QueryRowContext(ctx, eventCountQuery).Scan(&eventCount); err != nil {
		log.Fatalf("Failed to count events: %v", err)
	}

	fmt.Printf("Total events: %d\n\n", eventCount)

	if eventCount == 0 {
		log.Fatal("No events found. Run 'make ingest' first.")
	}

	// ========================================================================
	// STEP 2: Refresh views
	// ========================================================================
	fmt.Println("Step 2: Refreshing materialized views...\n")

	views := []string{
		"daily_user_activity",
		"weekly_team_activity",
		"user_correlation_summary",
	}

	for _, viewName := range views {
		refreshWriter := RefreshView(ctx, conn, viewName)
		refreshRes, refreshLogs := refreshWriter.Run()

		for _, log := range refreshLogs {
			fmt.Printf("  %s\n", log)
		}

		if !refreshRes.IsOk() {
			fmt.Printf("  Warning: %v\n", refreshRes.Error())
		}
	}
	fmt.Println()

	// ========================================================================
	// STEP 3: Check view row counts
	// ========================================================================
	fmt.Println("Step 3: Checking materialized view data...\n")

	for _, viewName := range views {
		checkWriter := CheckViewData(ctx, conn, viewName)
		checkRes, checkLogs := checkWriter.Run()

		for _, log := range checkLogs {
			fmt.Printf("  %s\n", log)
		}

		if checkRes.IsOk() {
			count := checkRes.Unwrap()
			fmt.Printf("  Rows in %s: %d\n\n", viewName, count)
		} else {
			fmt.Printf("  Error: %v\n\n", checkRes.Error())
		}
	}

	// ========================================================================
	// STEP 4: Fetch daily activity (should have data)
	// ========================================================================
	fmt.Println("Step 4: Fetching daily user activity using monoids...\n")

	dailyWriter := FetchDailyUserActivityData(ctx, conn, 5)
	dailyRes, dailyLogs := dailyWriter.Run()

	for _, log := range dailyLogs {
		fmt.Printf("  %s\n", log)
	}

	if dailyRes.IsOk() {
		dailyData := dailyRes.Unwrap()
		fmt.Printf("\nFound %d daily activity rows:\n", len(dailyData))

		for i, row := range dailyData {
			fmt.Printf("  [%d] User: %s | Platform: %s | Events: %d | Date: %s\n",
				i+1, row.UserID, row.Source, row.EventCount, row.Date.Format("2006-01-02"))
		}
	} else {
		fmt.Printf("Error: %v\n", dailyRes.Error())
	}
	fmt.Println()

	// ========================================================================
	// STEP 5: Fetch correlations (may be empty)
	// ========================================================================
	fmt.Println("Step 5: Fetching user correlations using monoids...\n")

	corrWriter := FetchCorrelationData(ctx, conn, 5)
	corrRes, corrLogs := corrWriter.Run()

	for _, log := range corrLogs {
		fmt.Printf("  %s\n", log)
	}

	if corrRes.IsOk() {
		corrData := corrRes.Unwrap()
		fmt.Printf("\nFound %d correlation rows:\n", len(corrData))

		if len(corrData) == 0 {
			fmt.Println("  (No correlations detected yet - correlation detector may not have run)")
		} else {
			for i, row := range corrData {
				fmt.Printf("  [%d] User: %s | Type: %s | Confidence: %.2f | Frequency: %d\n",
					i+1, row.UserID, row.CorrelationType, row.AvgConfidence, row.TotalFrequency)
			}
		}
	} else {
		fmt.Printf("Error: %v\n", corrRes.Error())
	}
	fmt.Println()

	// ========================================================================
	// STEP 6: Fetch team activity (may be empty)
	// ========================================================================
	fmt.Println("Step 6: Fetching weekly team activity using monoids...\n")

	teamWriter := FetchTeamActivityData(ctx, conn, 5)
	teamRes, teamLogs := teamWriter.Run()

	for _, log := range teamLogs {
		fmt.Printf("  %s\n", log)
	}

	if teamRes.IsOk() {
		teamData := teamRes.Unwrap()
		fmt.Printf("\nFound %d team activity rows:\n", len(teamData))

		if len(teamData) == 0 {
			fmt.Println("  (No team activity rows - view may not be grouping data correctly)")
		} else {
			for i, row := range teamData {
				fmt.Printf("  [%d] Team: %s | Week: %s | Platform: %s | Active Members: %d\n",
					i+1, row.TeamID, row.WeekStart.Format("2006-01-02"), row.Source, row.ActiveMembers)
			}
		}
	} else {
		fmt.Printf("Error: %v\n", teamRes.Error())
	}
	fmt.Println()

	// ========================================================================
	// STEP 7: Test Claude analysis with real activity
	// ========================================================================
	fmt.Println("Step 7: Testing Claude weekly analysis...\n")

	if dailyRes.IsOk() && len(dailyRes.Unwrap()) > 0 {
		dailyData := dailyRes.Unwrap()
		firstUserID := dailyData[0].UserID

		fmt.Printf("Analyzing week for user: %s\n\n", firstUserID)

		// Create Claude client
		// cmd/test-weekly-metrics/main.go

		claudeClient := &llm.ClaudeClient{
			APIKey:     os.Getenv("ANTHROPIC_API_KEY"),
			Model:      "claude-sonnet-4-20250514",
			BaseURL:    "https://api.anthropic.com/v1", // Include /v1
			MaxRetries: 3,
			Timeout:    60 * time.Second,
		}

		// Analyze activity
		week := types.TimeRange{
			Start: time.Now().AddDate(0, 0, -7),
			End:   time.Now(),
		}

		analysisWriter := llm.AnalyzeWeeklyActivity(ctx, claudeClient, firstUserID, week, dailyData)
		analysisRes, analysisLogs := analysisWriter.Run()

		fmt.Println("Analysis logs:")
		for _, log := range analysisLogs {
			fmt.Printf("  %s\n", log)
		}

		if analysisRes.IsOk() {
			report := analysisRes.Unwrap()
			fmt.Printf("\n✅ Weekly Report Generated:\n")
			fmt.Printf("  Wins: %d\n", len(report.Wins))
			for _, w := range report.Wins {
				fmt.Printf("    - %s: %s\n", w.Title, w.Impact)
			}
			fmt.Printf("  In Progress: %d\n", len(report.InProgress))
			for _, p := range report.InProgress {
				fmt.Printf("    - %s (%d%% complete)\n", p.Title, p.PercentComplete)
			}
			fmt.Printf("  Blocked: %d\n", len(report.Blocked))
			for _, b := range report.Blocked {
				fmt.Printf("    - %s (blocked by: %s)\n", b.Title, b.BlockedBy)
			}
		} else {
			fmt.Printf("  Error: %v\n", analysisRes.Error())
		}
	}

	fmt.Println("\nDiagnostic complete!")
}
