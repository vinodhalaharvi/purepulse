package analytics

import (
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// CORRELATION
// ============================================================================

// Correlation represents detected cross-platform pattern
type Correlation struct {
	ID          types.CorrelationID   `json:"id"`
	UserID      types.UserID          `json:"user_id"`
	Type        types.CorrelationType `json:"type"`
	Events      []events.Event        `json:"events"`     // Related events (source + target)
	Confidence  float64               `json:"confidence"` // 0.0 - 1.0
	Description string                `json:"description"`
	TimeDelta   time.Duration         `json:"time_delta"` // Time between correlated events
	Frequency   int                   `json:"frequency"`  // How many times pattern occurred
	DetectedAt  time.Time             `json:"detected_at"`
}

// IsHighConfidence checks if correlation has high confidence (>= 0.7)
func (c Correlation) IsHighConfidence() bool {
	return c.Confidence >= 0.7
}

// IsMediumConfidence checks if correlation has medium confidence (0.5 - 0.7)
func (c Correlation) IsMediumConfidence() bool {
	return c.Confidence >= 0.5 && c.Confidence < 0.7
}

// IsLowConfidence checks if correlation has low confidence (< 0.5)
func (c Correlation) IsLowConfidence() bool {
	return c.Confidence < 0.5
}

// TimeDeltaMinutes returns time delta in minutes
func (c Correlation) TimeDeltaMinutes() int {
	return int(c.TimeDelta.Minutes())
}

// TimeDeltaHours returns time delta in hours
func (c Correlation) TimeDeltaHours() float64 {
	return c.TimeDelta.Hours()
}

// ============================================================================
// CORRELATION WINDOW
// ============================================================================

// CorrelationWindow for time-window correlation
type CorrelationWindow struct {
	Duration time.Duration `json:"duration"` // e.g., 1 hour
	Offset   time.Duration `json:"offset"`   // Slide window by this amount
}

// DefaultCorrelationWindow returns default window (1 hour duration, 15 min offset)
func DefaultCorrelationWindow() CorrelationWindow {
	return CorrelationWindow{
		Duration: 1 * time.Hour,
		Offset:   15 * time.Minute,
	}
}
