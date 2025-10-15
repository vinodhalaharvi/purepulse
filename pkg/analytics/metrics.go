package analytics

import (
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// STRUCTURED METRICS
// ============================================================================

// StructuredMetrics are computed from raw events
type StructuredMetrics struct {
	UserID    types.UserID    `json:"user_id"`
	TimeRange types.TimeRange `json:"time_range"`

	// Slack metrics
	SlackMessages      int `json:"slack_messages"`
	SlackChannels      int `json:"slack_channels"`
	SlackReactions     int `json:"slack_reactions"`
	SlackFilesShared   int `json:"slack_files_shared"`
	SlackThreadReplies int `json:"slack_thread_replies"`

	// GitHub metrics
	GitHubCommits      int     `json:"github_commits"`
	GitHubPRs          int     `json:"github_prs"`
	GitHubPRsReviewed  int     `json:"github_prs_reviewed"`
	GitHubIssues       int     `json:"github_issues"`
	GitHubComments     int     `json:"github_comments"`
	GitHubLinesAdded   int     `json:"github_lines_added"`
	GitHubLinesDeleted int     `json:"github_lines_deleted"`
	CodeQuality        float64 `json:"code_quality"` // Derived score

	// Jira metrics
	JiraIssuesCreated   int     `json:"jira_issues_created"`
	JiraIssuesCompleted int     `json:"jira_issues_completed"`
	JiraStoryPoints     float64 `json:"jira_story_points"`
	JiraVelocity        float64 `json:"jira_velocity"` // Points per sprint
	JiraComments        int     `json:"jira_comments"`
	JiraTransitions     int     `json:"jira_transitions"`

	// Zoom metrics
	ZoomMeetings   int `json:"zoom_meetings"`
	ZoomMinutes    int `json:"zoom_minutes"`
	ZoomWebinars   int `json:"zoom_webinars"`
	ZoomRecordings int `json:"zoom_recordings"`

	// Derived cross-platform metrics
	ActiveHours        float64 `json:"active_hours"`
	PeakHour           int     `json:"peak_hour"`           // 0-23
	CollaborationScore float64 `json:"collaboration_score"` // 0-100

	// Metadata
	DataPoints int `json:"data_points"` // Total events processed
}

// TotalSlackActivity returns total Slack activity
func (sm StructuredMetrics) TotalSlackActivity() int {
	return sm.SlackMessages + sm.SlackReactions + sm.SlackFilesShared + sm.SlackThreadReplies
}

// TotalGitHubActivity returns total GitHub activity
func (sm StructuredMetrics) TotalGitHubActivity() int {
	return sm.GitHubCommits + sm.GitHubPRs + sm.GitHubIssues + sm.GitHubComments
}

// TotalJiraActivity returns total Jira activity
func (sm StructuredMetrics) TotalJiraActivity() int {
	return sm.JiraIssuesCreated + sm.JiraIssuesCompleted + sm.JiraComments + sm.JiraTransitions
}

// TotalZoomActivity returns total Zoom activity
func (sm StructuredMetrics) TotalZoomActivity() int {
	return sm.ZoomMeetings + sm.ZoomWebinars
}

// TotalActivity returns total activity across all platforms
func (sm StructuredMetrics) TotalActivity() int {
	return sm.TotalSlackActivity() + sm.TotalGitHubActivity() +
		sm.TotalJiraActivity() + sm.TotalZoomActivity()
}

// GitHubNetLines returns net lines changed (additions - deletions)
func (sm StructuredMetrics) GitHubNetLines() int {
	return sm.GitHubLinesAdded - sm.GitHubLinesDeleted
}

// ZoomHours returns Zoom time in hours
func (sm StructuredMetrics) ZoomHours() float64 {
	return float64(sm.ZoomMinutes) / 60.0
}
