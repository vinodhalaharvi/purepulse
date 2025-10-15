// Package monoids provides custom Monoid implementations for domain types.
//
// All monoids in this package implement the github.com/vinodhalaharvi/purekernels/pkg/monoid.Monoid interface
// and can be used with applicative functors for parallel composition.
//
// Available Monoids:
//   - UserActivityMonoid: Combines user activity from multiple sources
//   - FetchMetadataMonoid: Combines API fetch performance metrics
//   - FetchResultMonoid: Combines fetch results (activity + errors + metadata)
//   - StructuredMetricsMonoid: Combines analytical metrics using sum and average monoids
//   - AISummaryMonoid: Combines AI-generated summaries from multiple LLM calls
//   - UserSummaryMonoid: Combines complete user summaries
//   - TeamSummaryMonoid: Combines team-level summaries
//   - AggregatedTeamMetricsMonoid: Combines team-level metrics
//
// Usage Example:
//
//	import (
//	    "github.com/vinodhalaharvi/purekernels/pkg/functor"
//	    "github.com/vinodhalaharvi/purepulse/pkg/monoids"
//	    "github.com/vinodhalaharvi/purepulse/pkg/events"
//	)
//
//	// Create monoid instance
//	monoid := monoids.FetchResultMonoid{}
//
//	// Use with Concurrent applicative for parallel composition
//	fetch1 := functor.NewConcurrent(monoid, func() events.FetchResult {
//	    // Fetch from source 1
//	    return result1
//	})
//
//	fetch2 := functor.NewConcurrent(monoid, func() events.FetchResult {
//	    // Fetch from source 2
//	    return result2
//	})
//
//	// Combine results in parallel
//	combined := fetch1.Apply(fetch2).Value()
package monoids
