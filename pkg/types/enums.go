package types

// ============================================================================
// PLATFORM ENUM
// ============================================================================

// Platform represents supported integration platforms
type Platform string

const (
	PlatformSlack  Platform = "slack"
	PlatformGitHub Platform = "github"
	PlatformJira   Platform = "jira"
	PlatformZoom   Platform = "zoom"
)

// String returns string representation
func (p Platform) String() string {
	return string(p)
}

// IsValid checks if platform is valid
func (p Platform) IsValid() bool {
	switch p {
	case PlatformSlack, PlatformGitHub, PlatformJira, PlatformZoom:
		return true
	default:
		return false
	}
}

// ============================================================================
// EVENT TYPE ENUM
// ============================================================================

// EventType represents normalized event types across platforms
type EventType string

// Slack event types
const (
	EventSlackMessage     EventType = "slack_message"
	EventSlackReaction    EventType = "slack_reaction"
	EventSlackFileUpload  EventType = "slack_file_upload"
	EventSlackChannelJoin EventType = "slack_channel_join"
	EventSlackThreadReply EventType = "slack_thread_reply"
)

// GitHub event types
const (
	EventGitHubCommit       EventType = "github_commit"
	EventGitHubPR           EventType = "github_pull_request"
	EventGitHubPRReview     EventType = "github_pr_review"
	EventGitHubPRComment    EventType = "github_pr_comment"
	EventGitHubIssue        EventType = "github_issue"
	EventGitHubIssueComment EventType = "github_issue_comment"
	EventGitHubRelease      EventType = "github_release"
)

// Jira event types
const (
	EventJiraIssueCreated EventType = "jira_issue_created"
	EventJiraIssueUpdated EventType = "jira_issue_updated"
	EventJiraIssueClosed  EventType = "jira_issue_closed"
	EventJiraComment      EventType = "jira_comment"
	EventJiraTransition   EventType = "jira_transition"
	EventJiraSprint       EventType = "jira_sprint"
)

// Zoom event types
const (
	EventZoomMeeting   EventType = "zoom_meeting"
	EventZoomWebinar   EventType = "zoom_webinar"
	EventZoomRecording EventType = "zoom_recording"
	EventZoomChat      EventType = "zoom_chat"
)

// String returns string representation
func (e EventType) String() string {
	return string(e)
}

// Platform returns the platform this event type belongs to
func (e EventType) Platform() Platform {
	switch e {
	case EventSlackMessage, EventSlackReaction, EventSlackFileUpload,
		EventSlackChannelJoin, EventSlackThreadReply:
		return PlatformSlack
	case EventGitHubCommit, EventGitHubPR, EventGitHubPRReview,
		EventGitHubPRComment, EventGitHubIssue, EventGitHubIssueComment,
		EventGitHubRelease:
		return PlatformGitHub
	case EventJiraIssueCreated, EventJiraIssueUpdated, EventJiraIssueClosed,
		EventJiraComment, EventJiraTransition, EventJiraSprint:
		return PlatformJira
	case EventZoomMeeting, EventZoomWebinar, EventZoomRecording, EventZoomChat:
		return PlatformZoom
	default:
		return ""
	}
}

// ============================================================================
// CORRELATION TYPE ENUM
// ============================================================================

// CorrelationType represents detected cross-platform patterns
type CorrelationType string

const (
	CorrelationSlackToGitHub  CorrelationType = "slack_to_github"
	CorrelationJiraToGitHub   CorrelationType = "jira_to_github"
	CorrelationZoomToActivity CorrelationType = "zoom_to_activity"
	CorrelationGitHubToJira   CorrelationType = "github_to_jira"
	CorrelationSlackToJira    CorrelationType = "slack_to_jira"
)

// String returns string representation
func (c CorrelationType) String() string {
	return string(c)
}

// ============================================================================
// HTTP STATUS CODE ENUM
// ============================================================================

// StatusCode represents HTTP status codes
type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusCreated             StatusCode = 201
	StatusAccepted            StatusCode = 202
	StatusBadRequest          StatusCode = 400
	StatusUnauthorized        StatusCode = 401
	StatusForbidden           StatusCode = 403
	StatusNotFound            StatusCode = 404
	StatusTooManyRequests     StatusCode = 429
	StatusInternalServerError StatusCode = 500
	StatusServiceUnavailable  StatusCode = 503
)

// String returns string representation
func (s StatusCode) String() string {
	switch s {
	case StatusOK:
		return "200 OK"
	case StatusCreated:
		return "201 Created"
	case StatusAccepted:
		return "202 Accepted"
	case StatusBadRequest:
		return "400 Bad Request"
	case StatusUnauthorized:
		return "401 Unauthorized"
	case StatusForbidden:
		return "403 Forbidden"
	case StatusNotFound:
		return "404 Not Found"
	case StatusTooManyRequests:
		return "429 Too Many Requests"
	case StatusInternalServerError:
		return "500 Internal Server Error"
	case StatusServiceUnavailable:
		return "503 Service Unavailable"
	default:
		return "Unknown Status"
	}
}

// IsSuccess returns true if status is 2xx
func (s StatusCode) IsSuccess() bool {
	return s >= 200 && s < 300
}

// IsError returns true if status is 4xx or 5xx
func (s StatusCode) IsError() bool {
	return s >= 400
}
