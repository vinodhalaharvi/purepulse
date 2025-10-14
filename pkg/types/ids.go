package types

// ============================================================================
// IDENTIFIERS
// ============================================================================

// UserID identifies a user across platforms
type UserID string

// String returns string representation
func (u UserID) String() string {
	return string(u)
}

// IsEmpty checks if UserID is empty
func (u UserID) IsEmpty() bool {
	return string(u) == ""
}

// TeamID identifies a team
type TeamID string

// String returns string representation
func (t TeamID) String() string {
	return string(t)
}

// IsEmpty checks if TeamID is empty
func (t TeamID) IsEmpty() bool {
	return string(t) == ""
}

// EventID identifies an event
type EventID string

// String returns string representation
func (e EventID) String() string {
	return string(e)
}

// CorrelationID identifies a correlation
type CorrelationID string

// String returns string representation
func (c CorrelationID) String() string {
	return string(c)
}
