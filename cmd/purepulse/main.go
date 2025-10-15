package main

import (
	"context"
	"fmt"
	"time"

	"github.com/vinodhalaharvi/purepulse/internal/connectors"
	"github.com/vinodhalaharvi/purepulse/pkg/collectors"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

func main() {
	// 1. Create platform config
	config := types.PlatformConfig{
		Slack:  types.SlackConfig{Token: "mock-token"},
		GitHub: types.GitHubConfig{Token: "mock-token"},
		Jira:   types.JiraConfig{URL: "mock-url"},
		Zoom:   types.ZoomConfig{AccountID: "mock-account"},
	}

	// 2. Create LOW-LEVEL connectors (mock implementations)
	connectorSet := connectors.NewMockConnectors(config)
	connectorsMap := connectorSet.AsMap()

	// 3. Use HIGH-LEVEL collector (applicative composition)
	ctx := context.Background()
	userID := types.UserID("alice")
	timeRange := types.TimeRange{
		Start: time.Now().AddDate(0, 0, -30),
		End:   time.Now(),
	}

	// Validate input (Validation applicative)
	validation := collectors.ValidateFetchRequest(userID, timeRange)
	if !validation.IsValid() {
		fmt.Printf("Validation failed: %v\n", validation.GetErrors())
		return
	}

	// Fetch in parallel with audit logging (Concurrent + Writer applicatives)
	auditedFetch := collectors.FetchAllPlatformsAudited(
		ctx,
		connectorsMap, // ← LOW-LEVEL connectors passed to HIGH-LEVEL collector
		userID,
		timeRange,
	)

	// Execute and get result + logs
	result, logs := auditedFetch.Run()

	// Print results
	fmt.Printf("Total events: %d\n", result.Activity.TotalEvents())
	fmt.Printf("Errors: %d\n", len(result.Errors))
	fmt.Printf("Audit log:\n")
	for _, log := range logs {
		fmt.Printf("  - %s\n", log)
	}
}
