package collectors

import (
	"context"
	"fmt"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/effect"
	"github.com/vinodhalaharvi/purepulse/internal/connectors"
	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// AUDITED FETCH (Writer Applicative)
// ============================================================================

// FetchAllPlatformsAudited wraps fetch with audit logging
func FetchAllPlatformsAudited(
	ctx context.Context,
	connectors map[types.Platform]connectors.Connector,
	userID types.UserID,
	timeRange types.TimeRange,
) effect.Writer[[]string, events.FetchResult] {

	start := time.Now()

	// Generate initial log
	logs := []string{
		fmt.Sprintf("fetch_started: user=%s, platforms=%d, time_range=%s",
			userID,
			len(connectors),
			timeRange.Duration(),
		),
	}

	// Execute parallel fetch
	concurrent := FetchAllPlatforms(ctx, connectors, userID, timeRange)
	result := concurrent.Value() // Execute all in parallel

	elapsed := time.Since(start)

	// Generate completion logs
	logs = append(logs,
		fmt.Sprintf("fetch_completed: %d sources in %v", len(connectors), elapsed),
		fmt.Sprintf("total_events: %d", result.Activity.TotalEvents()),
		fmt.Sprintf("errors: %d", len(result.Errors)),
		fmt.Sprintf("total_api_calls: %d", result.Metadata.TotalAPICalls),
		fmt.Sprintf("cache_hit_rate: %.2f%%", result.Metadata.CacheHitRate()),
	)

	// Add per-platform logs
	for platform, metrics := range result.Metadata.SourceMetrics {
		if metrics.Success {
			logs = append(logs,
				fmt.Sprintf("platform=%s: %d events in %dms",
					platform, metrics.Events, metrics.LatencyMS),
			)
		} else {
			logs = append(logs,
				fmt.Sprintf("platform=%s: FAILED - %s",
					platform, metrics.Error),
			)
		}
	}

	return effect.NewWriter(result, logs)
}
