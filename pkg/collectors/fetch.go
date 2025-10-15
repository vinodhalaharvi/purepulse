package collectors

import (
	"context"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/functor"
	"github.com/vinodhalaharvi/purepulse/internal/connectors"
	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/monoids"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// APPLICATIVE COMPOSITION - PARALLEL FETCH WITH AUDIT
// ============================================================================

// FetchAllPlatforms fetches from all platforms in parallel using Concurrent applicative
func FetchAllPlatforms(
	ctx context.Context,
	connectors map[types.Platform]connectors.Connector,
	userID types.UserID,
	timeRange types.TimeRange,
) functor.Concurrent[events.FetchResult] {

	monoid := monoids.FetchResultMonoid{}

	// Create concurrent computation for each platform
	var computations []functor.Concurrent[events.FetchResult]

	for platform, connector := range connectors {
		// Capture in closure
		p := platform
		c := connector

		computation := functor.NewConcurrent(monoid, func() events.FetchResult {
			return fetchSinglePlatform(ctx, c, userID, timeRange, p)
		})

		computations = append(computations, computation)
	}

	// Combine all computations using Concurrent applicative
	if len(computations) == 0 {
		return functor.NewConcurrent(monoid, func() events.FetchResult {
			return monoid.Empty()
		})
	}

	// Start with first computation
	combined := computations[0]

	// Apply rest
	for i := 1; i < len(computations); i++ {
		combined = combined.Apply(computations[i])
	}

	return combined
}

// fetchSinglePlatform fetches from one platform (pure wrapper around impure I/O)
func fetchSinglePlatform(
	ctx context.Context,
	connector connectors.Connector,
	userID types.UserID,
	timeRange types.TimeRange,
	platform types.Platform,
) events.FetchResult {

	start := time.Now()

	// Call impure I/O at boundary
	fetchedEvents, err := connector.Fetch(ctx, userID, timeRange)

	latencyMS := int(time.Since(start).Milliseconds())

	// Build FetchResult (pure)
	if err != nil {
		// Error case: return empty activity + error
		return events.FetchResult{
			Activity: monoids.UserActivityMonoid{}.Empty(),
			Errors: []events.FetchError{{
				Source:    platform,
				Error:     err,
				Timestamp: time.Now(),
				Retryable: isRetryable(err),
			}},
			Metadata: buildErrorMetadata(platform, latencyMS, err),
		}
	}

	// Success case: return activity
	activity := buildUserActivity(userID, timeRange, platform, fetchedEvents)
	metadata := buildSuccessMetadata(platform, len(fetchedEvents), latencyMS)

	return events.FetchResult{
		Activity: activity,
		Errors:   []events.FetchError{},
		Metadata: metadata,
	}
}

// ============================================================================
// HELPER: BUILD USER ACTIVITY
// ============================================================================

func buildUserActivity(
	userID types.UserID,
	timeRange types.TimeRange,
	platform types.Platform,
	fetchedEvents []events.Event,
) events.UserActivity {

	activity := monoids.UserActivityMonoid{}.Empty()
	activity.UserID = userID
	activity.TimeRange = timeRange

	// Assign events to correct platform field
	switch platform {
	case types.PlatformSlack:
		activity.Slack = fetchedEvents
	case types.PlatformGitHub:
		activity.GitHub = fetchedEvents
	case types.PlatformJira:
		activity.Jira = fetchedEvents
	case types.PlatformZoom:
		activity.Zoom = fetchedEvents
	}

	return activity
}

// ============================================================================
// HELPER: BUILD METADATA
// ============================================================================

func buildSuccessMetadata(
	platform types.Platform,
	eventCount int,
	latencyMS int,
) events.FetchMetadata {

	sourceMetrics := map[types.Platform]events.SourceMetrics{
		platform: {
			APICalls:  1,
			Events:    eventCount,
			LatencyMS: latencyMS,
			Success:   true,
		},
	}

	return events.FetchMetadata{
		TotalAPICalls:  1,
		TotalEvents:    eventCount,
		TotalLatencyMS: latencyMS,
		CacheHits:      0,
		FetchedAt:      time.Now(),
		SourceMetrics:  sourceMetrics,
	}
}

func buildErrorMetadata(
	platform types.Platform,
	latencyMS int,
	err error,
) events.FetchMetadata {

	sourceMetrics := map[types.Platform]events.SourceMetrics{
		platform: {
			APICalls:  1,
			Events:    0,
			LatencyMS: latencyMS,
			Success:   false,
			Error:     err.Error(),
		},
	}

	return events.FetchMetadata{
		TotalAPICalls:  1,
		TotalEvents:    0,
		TotalLatencyMS: latencyMS,
		CacheHits:      0,
		FetchedAt:      time.Now(),
		SourceMetrics:  sourceMetrics,
	}
}

// ============================================================================
// HELPER: ERROR CLASSIFICATION
// ============================================================================

func isRetryable(err error) bool {
	// TODO: Implement based on error type
	// - Rate limit errors: false (wait)
	// - Network errors: true
	// - Auth errors: false
	return false
}
