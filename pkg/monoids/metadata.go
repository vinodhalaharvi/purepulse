package monoids

import (
	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// FETCH METADATA MONOID
// ============================================================================

// FetchMetadataMonoid combines fetch performance metadata
type FetchMetadataMonoid struct{}

// Empty returns the identity element for FetchMetadata
func (FetchMetadataMonoid) Empty() events.FetchMetadata {
	return events.FetchMetadata{
		SourceMetrics: make(map[types.Platform]events.SourceMetrics),
	}
}

// Combine merges two FetchMetadata instances
func (m FetchMetadataMonoid) Combine(a, b events.FetchMetadata) events.FetchMetadata {
	// Use sum monoid for integers
	intSum := monoid.NewSumMonoid[int]()

	// Merge source metrics
	mergedMetrics := make(map[types.Platform]events.SourceMetrics)

	// Add metrics from a
	for platform, metrics := range a.SourceMetrics {
		mergedMetrics[platform] = metrics
	}

	// Merge or add metrics from b
	for platform, metricsB := range b.SourceMetrics {
		if metricsA, exists := mergedMetrics[platform]; exists {
			// Merge if both have metrics for same platform
			mergedMetrics[platform] = events.SourceMetrics{
				APICalls:    intSum.Combine(metricsA.APICalls, metricsB.APICalls),
				Events:      intSum.Combine(metricsA.Events, metricsB.Events),
				LatencyMS:   intSum.Combine(metricsA.LatencyMS, metricsB.LatencyMS),
				Success:     metricsA.Success && metricsB.Success,
				Cached:      metricsA.Cached || metricsB.Cached,
				RateLimited: metricsA.RateLimited || metricsB.RateLimited,
				Error:       coalesceString(metricsA.Error, metricsB.Error),
			}
		} else {
			mergedMetrics[platform] = metricsB
		}
	}

	// Use latest FetchedAt timestamp
	fetchedAt := a.FetchedAt
	if b.FetchedAt.After(a.FetchedAt) {
		fetchedAt = b.FetchedAt
	}

	return events.FetchMetadata{
		TotalAPICalls:  intSum.Combine(a.TotalAPICalls, b.TotalAPICalls),
		TotalEvents:    intSum.Combine(a.TotalEvents, b.TotalEvents),
		TotalLatencyMS: intSum.Combine(a.TotalLatencyMS, b.TotalLatencyMS),
		CacheHits:      intSum.Combine(a.CacheHits, b.CacheHits),
		FetchedAt:      fetchedAt,
		SourceMetrics:  mergedMetrics,
	}
}

// Verify FetchMetadataMonoid implements Monoid interface
var _ monoid.Monoid[events.FetchMetadata] = FetchMetadataMonoid{}

// ============================================================================
// FETCH RESULT MONOID
// ============================================================================

// FetchResultMonoid combines fetch results (activity + errors + metadata)
type FetchResultMonoid struct{}

// Empty returns the identity element for FetchResult
func (FetchResultMonoid) Empty() events.FetchResult {
	return events.FetchResult{
		Activity: UserActivityMonoid{}.Empty(),
		Errors:   []events.FetchError{},
		Metadata: FetchMetadataMonoid{}.Empty(),
	}
}

// Combine merges two FetchResult instances
func (m FetchResultMonoid) Combine(a, b events.FetchResult) events.FetchResult {
	activityMonoid := UserActivityMonoid{}
	metadataMonoid := FetchMetadataMonoid{}
	errorListMonoid := monoid.NewListMonoid[events.FetchError]()

	return events.FetchResult{
		Activity: activityMonoid.Combine(a.Activity, b.Activity),
		Errors:   errorListMonoid.Combine(a.Errors, b.Errors),
		Metadata: metadataMonoid.Combine(a.Metadata, b.Metadata),
	}
}

// Verify FetchResultMonoid implements Monoid interface
var _ monoid.Monoid[events.FetchResult] = FetchResultMonoid{}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// coalesceString returns first non-empty string
func coalesceString(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
