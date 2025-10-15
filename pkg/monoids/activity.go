package monoids

import (
	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// USER ACTIVITY MONOID
// ============================================================================

// UserActivityMonoid combines partial UserActivity results
type UserActivityMonoid struct{}

// Empty returns the identity element for UserActivity
func (UserActivityMonoid) Empty() events.UserActivity {
	return events.UserActivity{
		Slack:  []events.Event{},
		GitHub: []events.Event{},
		Jira:   []events.Event{},
		Zoom:   []events.Event{},
		Metadata: events.FetchMetadata{
			SourceMetrics: make(map[types.Platform]events.SourceMetrics),
		},
	}
}

// Combine merges two UserActivity instances
func (m UserActivityMonoid) Combine(a, b events.UserActivity) events.UserActivity {
	// Use list monoid for events
	eventListMonoid := monoid.NewListMonoid[events.Event]()

	// Use FetchMetadata monoid
	metadataMonoid := FetchMetadataMonoid{}

	return events.UserActivity{
		UserID:    coalesceUserID(a.UserID, b.UserID),
		TimeRange: coalesceTimeRange(a.TimeRange, b.TimeRange),
		Slack:     eventListMonoid.Combine(a.Slack, b.Slack),
		GitHub:    eventListMonoid.Combine(a.GitHub, b.GitHub),
		Jira:      eventListMonoid.Combine(a.Jira, b.Jira),
		Zoom:      eventListMonoid.Combine(a.Zoom, b.Zoom),
		Metadata:  metadataMonoid.Combine(a.Metadata, b.Metadata),
	}
}

// Verify UserActivityMonoid implements Monoid interface
var _ monoid.Monoid[events.UserActivity] = UserActivityMonoid{}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// coalesceUserID returns first non-empty UserID
func coalesceUserID(a, b types.UserID) types.UserID {
	if !a.IsEmpty() {
		return a
	}
	return b
}

// coalesceTimeRange returns first non-zero TimeRange
func coalesceTimeRange(a, b types.TimeRange) types.TimeRange {
	if !a.Start.IsZero() {
		return a
	}
	return b
}
