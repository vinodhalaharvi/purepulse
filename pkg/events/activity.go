package events

import (
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// USER ACTIVITY (Per-Platform Events)
// ============================================================================

// UserActivity combines events from all platforms
type UserActivity struct {
	UserID    types.UserID    `json:"user_id"`
	TimeRange types.TimeRange `json:"time_range"`

	// Per-platform events
	Slack  []Event `json:"slack"`
	GitHub []Event `json:"github"`
	Jira   []Event `json:"jira"`
	Zoom   []Event `json:"zoom"`

	// Metadata
	Metadata FetchMetadata `json:"metadata"`
}

// TotalEvents returns total event count across all platforms
func (ua UserActivity) TotalEvents() int {
	return len(ua.Slack) + len(ua.GitHub) + len(ua.Jira) + len(ua.Zoom)
}

// EventsByPlatform returns events for a specific platform
func (ua UserActivity) EventsByPlatform(platform types.Platform) []Event {
	switch platform {
	case types.PlatformSlack:
		return ua.Slack
	case types.PlatformGitHub:
		return ua.GitHub
	case types.PlatformJira:
		return ua.Jira
	case types.PlatformZoom:
		return ua.Zoom
	default:
		return []Event{}
	}
}

// HasActivity checks if user has any activity
func (ua UserActivity) HasActivity() bool {
	return ua.TotalEvents() > 0
}

// PlatformCounts returns event count per platform
func (ua UserActivity) PlatformCounts() map[types.Platform]int {
	return map[types.Platform]int{
		types.PlatformSlack:  len(ua.Slack),
		types.PlatformGitHub: len(ua.GitHub),
		types.PlatformJira:   len(ua.Jira),
		types.PlatformZoom:   len(ua.Zoom),
	}
}

// ============================================================================
// FETCH METADATA (API Performance Tracking)
// ============================================================================

// FetchMetadata tracks fetch performance
type FetchMetadata struct {
	TotalAPICalls  int                              `json:"total_api_calls"`
	TotalEvents    int                              `json:"total_events"`
	TotalLatencyMS int                              `json:"total_latency_ms"`
	CacheHits      int                              `json:"cache_hits"`
	FetchedAt      time.Time                        `json:"fetched_at"`
	SourceMetrics  map[types.Platform]SourceMetrics `json:"source_metrics"`
}

// AverageLatency returns average latency per API call
func (fm FetchMetadata) AverageLatency() int {
	if fm.TotalAPICalls == 0 {
		return 0
	}
	return fm.TotalLatencyMS / fm.TotalAPICalls
}

// CacheHitRate returns cache hit rate as percentage
func (fm FetchMetadata) CacheHitRate() float64 {
	if fm.TotalAPICalls == 0 {
		return 0
	}
	return float64(fm.CacheHits) / float64(fm.TotalAPICalls) * 100
}

// SourceMetrics for per-platform stats
type SourceMetrics struct {
	APICalls    int    `json:"api_calls"`
	Events      int    `json:"events"`
	LatencyMS   int    `json:"latency_ms"`
	Success     bool   `json:"success"`
	Cached      bool   `json:"cached"`
	RateLimited bool   `json:"rate_limited"`
	Error       string `json:"error,omitempty"`
}

// ============================================================================
// FETCH RESULT (Combined Activity + Errors)
// ============================================================================

// FetchResult combines activity, errors, and metadata
type FetchResult struct {
	Activity UserActivity  `json:"activity"`
	Errors   []FetchError  `json:"errors"`
	Metadata FetchMetadata `json:"metadata"`
}

// HasErrors checks if there were any fetch errors
func (fr FetchResult) HasErrors() bool {
	return len(fr.Errors) > 0
}

// IsPartialSuccess checks if some platforms succeeded
func (fr FetchResult) IsPartialSuccess() bool {
	return fr.Activity.HasActivity() && fr.HasErrors()
}

// IsCompleteSuccess checks if all platforms succeeded
func (fr FetchResult) IsCompleteSuccess() bool {
	return fr.Activity.HasActivity() && !fr.HasErrors()
}

// IsCompleteFailure checks if all platforms failed
func (fr FetchResult) IsCompleteFailure() bool {
	return !fr.Activity.HasActivity() && fr.HasErrors()
}

// FetchError represents connector failure
type FetchError struct {
	Source      types.Platform `json:"source"`
	Error       error          `json:"error"`
	Timestamp   time.Time      `json:"timestamp"`
	Retryable   bool           `json:"retryable"`
	StatusCode  int            `json:"status_code,omitempty"`
	RateLimited bool           `json:"rate_limited"`
	RetryAfter  *time.Duration `json:"retry_after,omitempty"`
}

// ShouldRetry checks if error should be retried
func (fe FetchError) ShouldRetry() bool {
	return fe.Retryable && !fe.RateLimited
}
