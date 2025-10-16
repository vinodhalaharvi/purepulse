package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"github.com/vinodhalaharvi/purepulse/db"
	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/llm"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

func main() {
	fmt.Println("🚀 Testing LLM End-to-End Pipeline")
	fmt.Println("===================================\n")

	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Check for API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ ANTHROPIC_API_KEY not set in environment")
	}

	ctx := context.Background()

	// ========================================================================
	// STEP 1: Connect to Database
	// ========================================================================
	fmt.Println("📊 Step 1: Connecting to database...")

	dbConfig, err := db.NewConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to create database config: %v", err)
	}

	conn, err := db.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ Connected to database\n")

	// ========================================================================
	// STEP 2: Fetch Events from Database
	// ========================================================================
	fmt.Println("📥 Step 2: Fetching events from database...")

	userID := types.UserID("alice")

	// Query events from last 7 days
	query := `
        SELECT id, user_id, source, type, timestamp, payload, 
               author, channel, thread_id, parent_id, 
               size, duration_seconds
        FROM events
        WHERE user_id = $1 
          AND timestamp >= NOW() - INTERVAL '7 days'
        ORDER BY timestamp DESC
        LIMIT 100
    `

	rows, err := conn.DB.QueryContext(ctx, query, userID)
	if err != nil {
		log.Fatalf("Failed to query events: %v", err)
	}
	defer rows.Close()

	eventsList := []events.Event{}
	for rows.Next() {
		var e events.Event
		var threadID, parentID *string

		err := rows.Scan(
			&e.ID, &e.UserID, &e.Source, &e.Type, &e.Timestamp, &e.Payload,
			&e.Metadata.Author, &e.Metadata.Channel, &threadID, &parentID,
			&e.Metadata.Size, &e.Metadata.DurationSeconds,
		)
		if err != nil {
			log.Printf("Warning: Failed to scan event: %v", err)
			continue
		}

		if threadID != nil {
			e.Metadata.ThreadID = *threadID
		}
		if parentID != nil {
			e.Metadata.ParentID = *parentID
		}

		eventsList = append(eventsList, e)
	}

	if len(eventsList) == 0 {
		log.Fatal("❌ No events found for user. Run data ingestion first: go run ./cmd/purepulse/main.go")
	}

	fmt.Printf("✅ Fetched %d events\n\n", len(eventsList))

	// ========================================================================
	// STEP 3: Create Claude Client
	// ========================================================================
	fmt.Println("🤖 Step 3: Creating Claude client...")

	claudeClient := &llm.ClaudeClient{
		APIKey:     apiKey,
		Model:      "claude-sonnet-4-20250514",
		BaseURL:    "https://api.anthropic.com",
		MaxRetries: 3,
		Timeout:    60 * time.Second,
	}

	fmt.Println("✅ Claude client ready\n")

	// ========================================================================
	// STEP 4: Generate Daily Summary
	// ========================================================================
	fmt.Println("📝 Step 4: Generating daily summary with Claude...")

	date := time.Now().AddDate(0, 0, -1) // Yesterday

	summaryGenerator := &llm.SummaryGenerator{
		Client:         claudeClient,
		ResponseParser: llm.ParseDailySummary,
		Validator:      llm.ValidateBasicResponse,
	}

	startTime := time.Now()
	writer := summaryGenerator.GenerateDailySummary(ctx, userID, eventsList, date)
	summaryResult, logs := writer.Run()
	latencyMS := int(time.Since(startTime).Milliseconds())

	// Print audit logs
	fmt.Println("📋 Audit Trail:")
	for _, logEntry := range logs {
		fmt.Printf("  - %s\n", logEntry)
	}
	fmt.Println()

	if !summaryResult.IsOk() {
		log.Fatalf("❌ Summary generation failed: %v", summaryResult.Error())
	}

	summary := summaryResult.Unwrap()
	fmt.Println("✅ Summary generated successfully!")
	fmt.Printf("  - Content length: %d chars\n", len(summary.AISummary.Content)) // ← Changed
	fmt.Printf("  - Highlights: %d items\n", len(summary.AISummary.Highlights))  // ← Changed
	fmt.Printf("  - Insights: %d items\n", len(summary.AISummary.Insights))      // ← Changed
	fmt.Printf("  - Tokens used: %d (in: %d, out: %d)\n",
		summary.AITokens, // ← Changed (just total now)
		0,                // ← Input tokens not stored separately in new schema
		0,                // ← Output tokens not stored separately in new schema
	)
	fmt.Printf("  - Latency: %dms\n\n", latencyMS)

	// ========================================================================
	// STEP 5: Save Summary to Database
	// ========================================================================
	fmt.Println("💾 Step 5: Saving summary to database...")

	// Prepare summary for DB (adapt to existing schema)
	insertQuery := `
    INSERT INTO summaries (
        user_id,
        time_range_start,
        time_range_end,
        activity,
        metrics,
        correlations,
        ai_summary,
        ai_model,
        ai_tokens,
        ai_latency_ms,
        audit_log,
        version
    ) VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6::jsonb, $7::jsonb, $8, $9, $10, $11, $12)
    RETURNING id
`

	// Convert to JSON strings (PostgreSQL will cast to JSONB)
	activityJSON, _ := json.Marshal(map[string]interface{}{
		"total_events": len(eventsList),
		"slack_count":  countEventsByPlatform(eventsList, types.PlatformSlack),
		"github_count": countEventsByPlatform(eventsList, types.PlatformGitHub),
		"jira_count":   countEventsByPlatform(eventsList, types.PlatformJira),
		"zoom_count":   countEventsByPlatform(eventsList, types.PlatformZoom),
	})

	metricsJSON, _ := json.Marshal(map[string]interface{}{
		"total_events":        summary.Metrics.TotalEvents,
		"platforms_active":    summary.Metrics.PlatformsActive,
		"productivity_score":  summary.Metrics.ProductivityScore,
		"focus_score":         summary.Metrics.FocusScore,
		"collaboration_score": summary.Metrics.CollaborationScore,
	})

	correlationsJSON, _ := json.Marshal([]interface{}{})

	aiSummaryJSON, _ := json.Marshal(map[string]interface{}{
		"type":       summary.AISummary.Type,
		"content":    summary.AISummary.Content,
		"highlights": summary.AISummary.Highlights,
		"insights":   summary.AISummary.Insights,
	})

	var summaryID string
	err = conn.DB.QueryRowContext(ctx, insertQuery,
		userID,
		summary.TimeRangeStart,
		summary.TimeRangeEnd,
		string(activityJSON),     // ← Pass as JSON string
		string(metricsJSON),      // ← Pass as JSON string
		string(correlationsJSON), // ← Pass as JSON string
		string(aiSummaryJSON),    // ← Pass as JSON string
		summary.AIModel,
		summary.AITokens,
		latencyMS,
		pq.Array(logs), // ← Use pq.Array() for TEXT[]
		"1.0",
	).Scan(&summaryID)

	if err != nil {
		log.Fatalf("❌ Failed to save summary: %v", err)
	}

	fmt.Printf("✅ Summary saved with ID: %s\n\n", summaryID)

	// ========================================================================
	// STEP 6: Verify - Read Back from Database
	// ========================================================================
	fmt.Println("🔍 Step 6: Verifying saved summary...")

	verifyQuery := `
        SELECT 
            id, user_id, time_range_start, time_range_end,
            ai_summary->>'content' as content,
            jsonb_array_length(ai_summary->'highlights') as highlight_count,
            jsonb_array_length(ai_summary->'insights') as insight_count,
            ai_model, ai_tokens, ai_latency_ms
        FROM summaries
        WHERE id = $1
    `

	var (
		id, savedUserID, aiModel                      string
		savedStart, savedEnd                          time.Time
		savedContent                                  string
		highlightCount, insightCount, tokens, latency int
	)

	err = conn.DB.QueryRowContext(ctx, verifyQuery, summaryID).Scan(
		&id, &savedUserID, &savedStart, &savedEnd,
		&savedContent, &highlightCount, &insightCount,
		&aiModel, &tokens, &latency,
	)

	if err != nil {
		log.Fatalf("❌ Failed to verify summary: %v", err)
	}

	fmt.Println("✅ Summary verified in database!")
	fmt.Printf("  - ID: %s\n", id)
	fmt.Printf("  - User: %s\n", savedUserID)
	fmt.Printf("  - Time Range: %s to %s\n",
		savedStart.Format("2006-01-02"),
		savedEnd.Format("2006-01-02"),
	)
	fmt.Printf("  - Content: %d chars\n", len(savedContent))
	fmt.Printf("  - Highlights: %d\n", highlightCount)
	fmt.Printf("  - Insights: %d\n", insightCount)
	fmt.Printf("  - Model: %s\n", aiModel)
	fmt.Printf("  - Tokens: %d\n", tokens)
	fmt.Printf("  - Latency: %dms\n\n", latency)

	// ========================================================================
	// DONE!
	// ========================================================================
	fmt.Println("✨ End-to-End Test Complete!")
	fmt.Println("\n📊 Summary:")
	fmt.Printf("  ✅ Fetched %d events from database\n", len(eventsList))
	fmt.Printf("  ✅ Generated summary with Claude API\n")
	fmt.Printf("  ✅ Saved summary to database (ID: %s)\n", summaryID)
	fmt.Printf("  ✅ Verified summary is readable\n")
	fmt.Println("\n🎉 All systems working!")
}

// Helper function
func countEventsByPlatform(events []events.Event, platform types.Platform) int {
	count := 0
	for _, e := range events {
		if e.Source == platform {
			count++
		}
	}
	return count
}
