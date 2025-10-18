package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/vinodhalaharvi/purepulse/db"
	"github.com/vinodhalaharvi/purepulse/internal/connectors"
	"github.com/vinodhalaharvi/purepulse/pkg/collectors"
	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// ========================================================================
	// STEP 1: Connect to Database
	// ========================================================================

	dbConfig, err := db.NewConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to create database config: %v", err)
	}

	conn, err := db.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	log.Println("✓ Connected to database")

	// Check database health
	ctx := context.Background()
	health := conn.Health(ctx)
	log.Printf("Database status: %s (latency: %v)", health.Status, health.Latency)

	// ========================================================================
	// STEP 2: Create Platform Config
	// ========================================================================

	config := types.PlatformConfig{
		Slack:  types.SlackConfig{Token: "mock-token", WorkspaceID: "mock-workspace"},
		GitHub: types.GitHubConfig{Token: "mock-token", Organization: "mock-org"},
		Jira:   types.JiraConfig{URL: "https://mock.atlassian.net", Email: "mock@example.com", APIToken: "mock-token"},
		Zoom:   types.ZoomConfig{AccountID: "mock-account", ClientID: "mock-client", ClientSecret: "mock-secret"},
	}

	// ========================================================================
	// STEP 3: Create Connectors (Mock Implementations)
	// ========================================================================

	connectorSet := connectors.NewMockConnectors(config)
	connectorsMap := connectorSet.AsMap()

	log.Printf("✓ Created %d mock connectors", len(connectorsMap))

	// ========================================================================
	// STEP 4: Define Users and Time Range
	// ========================================================================

	userIDs := []types.UserID{"alice", "bob", "charlie", "david", "emma"}
	timeRange := types.TimeRange{
		Start: time.Now().AddDate(0, 0, -30), // Last 30 days
		End:   time.Now(),
	}

	log.Printf("✓ Fetching data for %d users from %s to %s",
		len(userIDs),
		timeRange.Start.Format("2006-01-02"),
		timeRange.End.Format("2006-01-02"),
	)

	// ========================================================================
	// STEP 5: Fetch and Load Data for Each User
	// ========================================================================

	for _, userID := range userIDs {
		log.Printf("\n--- Processing user: %s ---", userID)

		// Validate input (Validation applicative)
		validation := collectors.ValidateFetchRequest(userID, timeRange)
		if !validation.IsValid() {
			log.Printf("❌ Validation failed for %s: %v", userID, validation.GetErrors())
			continue
		}

		// Fetch in parallel with audit logging (Concurrent + Writer applicatives)
		auditedFetch := collectors.FetchAllPlatformsAudited(
			ctx,
			connectorsMap,
			userID,
			timeRange,
		)

		// Execute and get result + logs
		fetchResult, auditLogs := auditedFetch.Run()

		// Print audit logs
		log.Println("Audit trail:")
		for _, logEntry := range auditLogs {
			log.Printf("  %s", logEntry)
		}

		// Check for errors
		if len(fetchResult.Errors) > 0 {
			log.Printf("⚠️  Partial failure - %d sources failed:", len(fetchResult.Errors))
			for _, fetchErr := range fetchResult.Errors {
				log.Printf("  - %s: %v", fetchErr.Source, fetchErr.Error)
			}
		}

		// Print summary
		log.Printf("✓ Fetched %d total events", fetchResult.Activity.TotalEvents())
		log.Printf("  - Slack: %d events", len(fetchResult.Activity.Slack))
		log.Printf("  - GitHub: %d events", len(fetchResult.Activity.GitHub))
		log.Printf("  - Jira: %d events", len(fetchResult.Activity.Jira))
		log.Printf("  - Zoom: %d events", len(fetchResult.Activity.Zoom))

		// ====================================================================
		// STEP 6: Flatten Events and Insert to Database
		// ====================================================================

		allEvents := flattenEvents(fetchResult.Activity)

		if len(allEvents) == 0 {
			log.Printf("⚠️  No events to insert for user %s", userID)
			continue
		}

		log.Printf("Inserting %d events to database...", len(allEvents))

		insertStart := time.Now()
		inserted, err := insertEvents(ctx, conn.DB, allEvents)
		insertDuration := time.Since(insertStart)

		if err != nil {
			log.Printf("❌ Failed to insert events: %v", err)
			continue
		}

		log.Printf("✓ Inserted %d events in %v", inserted, insertDuration)
	}

	// ========================================================================
	// STEP 7: Print Database Summary
	// ========================================================================

	log.Println("\n=== Database Summary ===")

	// Count total events
	var totalEvents int
	err = conn.DB.QueryRow("SELECT COUNT(*) FROM events").Scan(&totalEvents)
	if err != nil {
		log.Printf("Failed to count events: %v", err)
	} else {
		log.Printf("Total events in database: %d", totalEvents)
	}

	// Count by platform
	rows, err := conn.DB.Query(`
        SELECT source, COUNT(*) as count
        FROM events
        GROUP BY source
        ORDER BY count DESC
    `)
	if err != nil {
		log.Printf("Failed to query event counts: %v", err)
	} else {
		defer rows.Close()
		log.Println("\nEvents by platform:")
		for rows.Next() {
			var source string
			var count int
			rows.Scan(&source, &count)
			log.Printf("  %s: %d events", source, count)
		}
	}

	// Count by user
	rows, err = conn.DB.Query(`
        SELECT user_id, COUNT(*) as count
        FROM events
        GROUP BY user_id
        ORDER BY count DESC
    `)
	if err != nil {
		log.Printf("Failed to query user counts: %v", err)
	} else {
		defer rows.Close()
		log.Println("\nEvents by user:")
		for rows.Next() {
			var userID string
			var count int
			rows.Scan(&userID, &count)
			log.Printf("  %s: %d events", userID, count)
		}
	}

	log.Println("\n✅ Data ingestion complete!")
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// flattenEvents converts UserActivity to flat []Event slice
func flattenEvents(activity events.UserActivity) []events.Event {
	allEvents := []events.Event{}
	allEvents = append(allEvents, activity.Slack...)
	allEvents = append(allEvents, activity.GitHub...)
	allEvents = append(allEvents, activity.Jira...)
	allEvents = append(allEvents, activity.Zoom...)
	return allEvents
}

// insertEvents bulk inserts events to database
func insertEvents(ctx context.Context, db *sql.DB, eventList []events.Event) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO events (
            user_id,
            source,
            type,
            timestamp,
            payload,
            author,
            channel,
            thread_id,
            parent_id,
            size,
            duration_seconds
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT DO NOTHING  -- Add this to skip duplicates
    `)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	for i, event := range eventList {
		// Handle nullable fields
		var threadID, parentID *string
		if event.Metadata.ThreadID != "" {
			threadID = &event.Metadata.ThreadID
		}
		if event.Metadata.ParentID != "" {
			parentID = &event.Metadata.ParentID
		}

		result, err := stmt.ExecContext(
			ctx,
			event.UserID,
			event.Source,
			event.Type,
			event.Timestamp,
			event.Payload,
			event.Metadata.Author,
			event.Metadata.Channel,
			threadID,
			parentID,
			event.Metadata.Size,
			event.Metadata.DurationSeconds,
		)

		if err != nil {
			// Log the FIRST error with details
			log.Printf("ERROR on event %d (%s): %v", i, event.ID, err)
			log.Printf("  UserID: %s, Source: %s, Type: %s", event.UserID, event.Source, event.Type)
			log.Printf("  Payload: %s", string(event.Payload))
			return 0, fmt.Errorf("failed at event %s: %w", event.ID, err)
		}

		rowsAffected, _ := result.RowsAffected()
		inserted += int(rowsAffected)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return inserted, nil
}
