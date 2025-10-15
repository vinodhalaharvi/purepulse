package events

import (
	"encoding/json"
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// EVENT (Normalized Domain Event)
// ============================================================================

// Event is the normalized unit across all platforms
type Event struct {
	ID        types.EventID   `json:"id"`
	UserID    types.UserID    `json:"user_id"`
	Source    types.Platform  `json:"source"`
	Type      types.EventType `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"` // Original platform data
	Metadata  EventMetadata   `json:"metadata"`
}

// PlatformMatches checks if event is from given platform
func (e Event) PlatformMatches(platform types.Platform) bool {
	return e.Source == platform
}

// TypeMatches checks if event is of given type
func (e Event) TypeMatches(eventType types.EventType) bool {
	return e.Type == eventType
}

// InTimeRange checks if event is within time range
func (e Event) InTimeRange(tr types.TimeRange) bool {
	return tr.Contains(e.Timestamp)
}

// Age returns how old the event is
func (e Event) Age() time.Duration {
	return time.Since(e.Timestamp)
}
