package analytics

import (
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// USER SUMMARY
// ============================================================================

// UserSummary is the final output combining all analysis layers
type UserSummary struct {
	UserID    types.UserID    `json:"user_id"`
	TimeRange types.TimeRange `json:"time_range"`

	// All layers combined
	Activity     events.UserActivity `json:"activity"`
	Metrics      StructuredMetrics   `json:"metrics"`
	Correlations []Correlation       `json:"correlations"`
	AISummary    AISummary           `json:"ai_summary"`

	// Metadata
	GeneratedAt time.Time `json:"generated_at"`
	AuditLog    []string  `json:"audit_log"`
	Version     string    `json:"version"` // Schema version
}

// HasCorrelations checks if summary has any correlations
func (us UserSummary) HasCorrelations() bool {
	return len(us.Correlations) > 0
}

// HighConfidenceCorrelations returns correlations with confidence >= 0.7
func (us UserSummary) HighConfidenceCorrelations() []Correlation {
	var high []Correlation
	for _, c := range us.Correlations {
		if c.IsHighConfidence() {
			high = append(high, c)
		}
	}
	return high
}

// ============================================================================
// TEAM SUMMARY
// ============================================================================

// TeamSummary aggregates multiple users
type TeamSummary struct {
	TeamID      types.TeamID          `json:"team_id"`
	TimeRange   types.TimeRange       `json:"time_range"`
	Members     []UserSummary         `json:"members"`
	Aggregated  AggregatedTeamMetrics `json:"aggregated"`
	GeneratedAt time.Time             `json:"generated_at"`
}

// MemberCount returns number of team members
func (ts TeamSummary) MemberCount() int {
	return len(ts.Members)
}

// TotalActivity returns sum of all member activity
func (ts TeamSummary) TotalActivity() int {
	total := 0
	for _, member := range ts.Members {
		total += member.Activity.TotalEvents()
	}
	return total
}

// ============================================================================
// AGGREGATED TEAM METRICS
// ============================================================================

// AggregatedTeamMetrics for team-level insights
type AggregatedTeamMetrics struct {
	TotalMembers  int `json:"total_members"`
	ActiveMembers int `json:"active_members"`

	// Aggregate counts
	TotalMessages int `json:"total_messages"`
	TotalCommits  int `json:"total_commits"`
	TotalPRs      int `json:"total_prs"`
	TotalIssues   int `json:"total_issues"`
	TotalMeetings int `json:"total_meetings"`

	// Team scores
	TeamVelocity       float64 `json:"team_velocity"`
	CollaborationScore float64 `json:"collaboration_score"`
	CodeQualityAvg     float64 `json:"code_quality_avg"`

	// Top contributors
	TopContributors []TopContributor `json:"top_contributors"`
}

// TopContributor for leaderboard
type TopContributor struct {
	UserID    types.UserID   `json:"user_id"`
	Name      string         `json:"name"`
	Score     float64        `json:"score"`
	Breakdown map[string]int `json:"breakdown"` // {"commits": 50, "prs": 10}
}

// ============================================================================
// AI SUMMARY
// ============================================================================

// AISummary from LLM
type AISummary struct {
	Overview        string      `json:"overview"`
	Highlights      []string    `json:"highlights"`
	Recommendations []string    `json:"recommendations"`
	Metadata        LLMMetadata `json:"metadata"`
}

// HasRecommendations checks if AI summary has recommendations
func (as AISummary) HasRecommendations() bool {
	return len(as.Recommendations) > 0
}

// LLMMetadata for LLM call stats
type LLMMetadata struct {
	Model        string  `json:"model"`
	Tokens       int     `json:"tokens"`
	LatencyMS    int     `json:"latency_ms"`
	CachedPrompt bool    `json:"cached_prompt"`
	Cost         float64 `json:"cost"` // Estimated API cost
}

// ============================================================================
// LLM PROMPT
// ============================================================================

// Prompt for LLM input
type Prompt struct {
	Audience  string `json:"audience"` // "executive" | "technical" | "team"
	Context   string `json:"context"`  // Generated from metrics + correlations
	MaxTokens int    `json:"max_tokens"`
}

// ============================================================================
// METADATA
// ============================================================================

// Metadata for versioning and timestamps
type Metadata struct {
	Version       string    `json:"version"`
	SchemaVersion string    `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	GeneratedBy   string    `json:"generated_by"` // Service/user that generated
}
