package query

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// SCANNER FUNCTIONS - Map SQL Rows to Domain Types
// ============================================================================

// Scanner is a function that maps a SQL row to a type T
type Scanner[T any] func(*sql.Rows) (T, error)

// SingleScanner is a function that maps a single SQL row to type T
type SingleScanner[T any] func(*sql.Row) (T, error)

// ============================================================================
// EVENT SCANNERS
// ============================================================================

// scanEvent maps a row to events.Event
func scanEvent(rows *sql.Rows) (events.Event, error) {
	var e events.Event
	var payload []byte
	var durationSeconds sql.NullInt32

	err := rows.Scan(
		&e.ID,
		&e.UserID,
		&e.Source,
		&e.Type,
		&e.Timestamp,
		&payload,
		&e.Metadata.Author,
		&e.Metadata.Channel,
		&e.Metadata.ThreadID,
		&e.Metadata.ParentID,
		&e.Metadata.Size,
		&durationSeconds,
	)

	if err != nil {
		return e, err
	}

	// Handle nullable duration
	if durationSeconds.Valid {
		ds := int(durationSeconds.Int32)
		e.Metadata.DurationSeconds = &ds
	}

	// Parse JSON payload
	e.Payload = json.RawMessage(payload)

	return e, nil
}

// scanEventSingle maps a single row to events.Event
func scanEventSingle(row *sql.Row) (events.Event, error) {
	var e events.Event
	var payload []byte
	var durationSeconds sql.NullInt32

	err := row.Scan(
		&e.ID,
		&e.UserID,
		&e.Source,
		&e.Type,
		&e.Timestamp,
		&payload,
		&e.Metadata.Author,
		&e.Metadata.Channel,
		&e.Metadata.ThreadID,
		&e.Metadata.ParentID,
		&e.Metadata.Size,
		&durationSeconds,
	)

	if err != nil {
		return e, err
	}

	if durationSeconds.Valid {
		ds := int(durationSeconds.Int32)
		e.Metadata.DurationSeconds = &ds
	}

	e.Payload = json.RawMessage(payload)

	return e, nil
}

// ============================================================================
// PLATFORM ACTIVITY SCANNER
// ============================================================================

// PlatformActivity represents aggregated activity per platform
type PlatformActivity struct {
	Source       types.Platform `json:"source"`
	UserID       types.UserID   `json:"user_id"`
	EventCount   int            `json:"event_count"`
	FirstEventAt time.Time      `json:"first_event_at"`
	LastEventAt  time.Time      `json:"last_event_at"`
}

// scanPlatformActivity maps row to PlatformActivity
func scanPlatformActivity(rows *sql.Rows) (PlatformActivity, error) {
	var pa PlatformActivity

	err := rows.Scan(
		&pa.Source,
		&pa.UserID,
		&pa.EventCount,
		&pa.FirstEventAt,
		&pa.LastEventAt,
	)

	return pa, err
}

// ============================================================================
// TEAM MEMBER REPORT SCANNER
// ============================================================================

// TeamMemberReport represents team member activity summary
type TeamMemberReport struct {
	UserID           types.UserID `json:"user_id"`
	DisplayName      string       `json:"display_name"`
	Email            string       `json:"email"`
	Role             string       `json:"role"`
	TotalEvents      int          `json:"total_events"`
	SlackEvents      int          `json:"slack_events"`
	GitHubEvents     int          `json:"github_events"`
	JiraEvents       int          `json:"jira_events"`
	ZoomEvents       int          `json:"zoom_events"`
	CorrelationCount int          `json:"correlation_count"`
	AvgConfidence    float64      `json:"avg_confidence"`
}

// scanTeamMemberReport maps row to TeamMemberReport
func scanTeamMemberReport(rows *sql.Rows) (TeamMemberReport, error) {
	var tmr TeamMemberReport
	var avgConfidence sql.NullFloat64

	err := rows.Scan(
		&tmr.UserID,
		&tmr.DisplayName,
		&tmr.Email,
		&tmr.Role,
		&tmr.TotalEvents,
		&tmr.SlackEvents,
		&tmr.GitHubEvents,
		&tmr.JiraEvents,
		&tmr.ZoomEvents,
		&tmr.CorrelationCount,
		&avgConfidence,
	)

	if err != nil {
		return tmr, err
	}

	// Handle nullable avg_confidence
	if avgConfidence.Valid {
		tmr.AvgConfidence = avgConfidence.Float64
	}

	return tmr, nil
}

// ============================================================================
// USER SCANNERS
// ============================================================================

// User represents a user from the database
type User struct {
	ID          types.UserID `json:"id"`
	DisplayName string       `json:"display_name"`
	Email       string       `json:"email"`
	Timezone    string       `json:"timezone"`
	Active      bool         `json:"active"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// scanUser maps row to User
func scanUser(rows *sql.Rows) (User, error) {
	var u User

	err := rows.Scan(
		&u.ID,
		&u.DisplayName,
		&u.Email,
		&u.Timezone,
		&u.Active,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	return u, err
}

// scanUserSingle maps single row to User
func scanUserSingle(row *sql.Row) (User, error) {
	var u User

	err := row.Scan(
		&u.ID,
		&u.DisplayName,
		&u.Email,
		&u.Timezone,
		&u.Active,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	return u, err
}

// ============================================================================
// CORRELATION SCANNERS
// ============================================================================

// CorrelationRow represents a correlation from database
type CorrelationRow struct {
	ID               types.CorrelationID   `json:"id"`
	UserID           types.UserID          `json:"user_id"`
	Type             types.CorrelationType `json:"type"`
	SourceEventID    int64                 `json:"source_event_id"`
	TargetEventID    int64                 `json:"target_event_id"`
	Confidence       float64               `json:"confidence"`
	TimeDeltaSeconds int                   `json:"time_delta_seconds"`
	Frequency        int                   `json:"frequency"`
	Description      string                `json:"description"`
	DetectedAt       time.Time             `json:"detected_at"`
}

// scanCorrelation maps row to CorrelationRow
func scanCorrelation(rows *sql.Rows) (CorrelationRow, error) {
	var c CorrelationRow

	err := rows.Scan(
		&c.ID,
		&c.UserID,
		&c.Type,
		&c.SourceEventID,
		&c.TargetEventID,
		&c.Confidence,
		&c.TimeDeltaSeconds,
		&c.Frequency,
		&c.Description,
		&c.DetectedAt,
	)

	return c, err
}

// ============================================================================
// SUMMARY SCANNERS
// ============================================================================

// SummaryRow represents a cached summary
type SummaryRow struct {
	ID             string          `json:"id"`
	UserID         types.UserID    `json:"user_id"`
	TimeRangeStart time.Time       `json:"time_range_start"`
	TimeRangeEnd   time.Time       `json:"time_range_end"`
	Activity       json.RawMessage `json:"activity"`
	Metrics        json.RawMessage `json:"metrics"`
	Correlations   json.RawMessage `json:"correlations"`
	AISummary      json.RawMessage `json:"ai_summary"`
	GeneratedAt    time.Time       `json:"generated_at"`
	ExpiresAt      *time.Time      `json:"expires_at"`
}

// scanSummary maps row to SummaryRow
func scanSummary(rows *sql.Rows) (SummaryRow, error) {
	var s SummaryRow
	var expiresAt sql.NullTime

	err := rows.Scan(
		&s.ID,
		&s.UserID,
		&s.TimeRangeStart,
		&s.TimeRangeEnd,
		&s.Activity,
		&s.Metrics,
		&s.Correlations,
		&s.AISummary,
		&s.GeneratedAt,
		&expiresAt,
	)

	if err != nil {
		return s, err
	}

	if expiresAt.Valid {
		s.ExpiresAt = &expiresAt.Time
	}

	return s, nil
}

// ============================================================================
// TEAM SCANNERS
// ============================================================================

// Team represents a team from database
type Team struct {
	ID          types.TeamID `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// scanTeam maps row to Team
func scanTeam(rows *sql.Rows) (Team, error) {
	var t Team

	err := rows.Scan(
		&t.ID,
		&t.Name,
		&t.Description,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	return t, err
}

// ============================================================================
// GENERIC SCANNERS FOR SIMPLE TYPES
// ============================================================================

// scanInt scans a single integer
func scanInt(row *sql.Row) (int, error) {
	var n int
	err := row.Scan(&n)
	return n, err
}

// scanString scans a single string
func scanString(row *sql.Row) (string, error) {
	var s string
	err := row.Scan(&s)
	return s, err
}

// scanBool scans a single boolean
func scanBool(row *sql.Row) (bool, error) {
	var b bool
	err := row.Scan(&b)
	return b, err
}
