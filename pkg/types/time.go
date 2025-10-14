package types

import "time"

// ============================================================================
// TIME RANGE
// ============================================================================

// TimeRange represents a time window for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Duration returns the duration of the time range
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// Contains checks if a time is within the range
func (tr TimeRange) Contains(t time.Time) bool {
	return !t.Before(tr.Start) && t.Before(tr.End)
}

// IsValid checks if the time range is valid (Start before End)
func (tr TimeRange) IsValid() bool {
	return tr.Start.Before(tr.End)
}

// Overlaps checks if this time range overlaps with another
func (tr TimeRange) Overlaps(other TimeRange) bool {
	return tr.Start.Before(other.End) && other.Start.Before(tr.End)
}
