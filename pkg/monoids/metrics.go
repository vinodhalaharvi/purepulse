package monoids

import (
	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// STRUCTURED METRICS MONOID
// ============================================================================

// StructuredMetricsMonoid combines analytical metrics
type StructuredMetricsMonoid struct{}

// Empty returns the identity element for StructuredMetrics
func (StructuredMetricsMonoid) Empty() analytics.StructuredMetrics {
	return analytics.StructuredMetrics{}
}

// Combine merges two StructuredMetrics instances
func (m StructuredMetricsMonoid) Combine(a, b analytics.StructuredMetrics) analytics.StructuredMetrics {
	// Use sum monoid for counts
	intSum := monoid.NewSumMonoid[int]()

	// Use average monoid for quality scores
	avgMonoid := monoid.NewAvgMonoid()

	// Combine code quality as weighted average
	codeQualityA := avgMonoid.FromValue(a.CodeQuality, a.GitHubCommits)
	codeQualityB := avgMonoid.FromValue(b.CodeQuality, b.GitHubCommits)
	combinedCodeQuality := avgMonoid.Combine(codeQualityA, codeQualityB)

	// Combine Jira velocity as weighted average
	velocityA := avgMonoid.FromValue(a.JiraVelocity, a.DataPoints)
	velocityB := avgMonoid.FromValue(b.JiraVelocity, b.DataPoints)
	combinedVelocity := avgMonoid.Combine(velocityA, velocityB)

	// Combine collaboration score as weighted average
	collabA := avgMonoid.FromValue(a.CollaborationScore, a.DataPoints)
	collabB := avgMonoid.FromValue(b.CollaborationScore, b.DataPoints)
	combinedCollab := avgMonoid.Combine(collabA, collabB)

	// Combine active hours as sum
	activeHoursSum := a.ActiveHours + b.ActiveHours

	// Peak hour: use the one with more activity
	peakHour := a.PeakHour
	if b.TotalActivity() > a.TotalActivity() {
		peakHour = b.PeakHour
	}

	return analytics.StructuredMetrics{
		UserID:    coalesceUserID(a.UserID, b.UserID),
		TimeRange: coalesceTimeRange(a.TimeRange, b.TimeRange),

		// Slack metrics (sum)
		SlackMessages:      intSum.Combine(a.SlackMessages, b.SlackMessages),
		SlackChannels:      intSum.Combine(a.SlackChannels, b.SlackChannels),
		SlackReactions:     intSum.Combine(a.SlackReactions, b.SlackReactions),
		SlackFilesShared:   intSum.Combine(a.SlackFilesShared, b.SlackFilesShared),
		SlackThreadReplies: intSum.Combine(a.SlackThreadReplies, b.SlackThreadReplies),

		// GitHub metrics (sum)
		GitHubCommits:      intSum.Combine(a.GitHubCommits, b.GitHubCommits),
		GitHubPRs:          intSum.Combine(a.GitHubPRs, b.GitHubPRs),
		GitHubPRsReviewed:  intSum.Combine(a.GitHubPRsReviewed, b.GitHubPRsReviewed),
		GitHubIssues:       intSum.Combine(a.GitHubIssues, b.GitHubIssues),
		GitHubComments:     intSum.Combine(a.GitHubComments, b.GitHubComments),
		GitHubLinesAdded:   intSum.Combine(a.GitHubLinesAdded, b.GitHubLinesAdded),
		GitHubLinesDeleted: intSum.Combine(a.GitHubLinesDeleted, b.GitHubLinesDeleted),
		CodeQuality:        combinedCodeQuality.Value(),

		// Jira metrics (sum)
		JiraIssuesCreated:   intSum.Combine(a.JiraIssuesCreated, b.JiraIssuesCreated),
		JiraIssuesCompleted: intSum.Combine(a.JiraIssuesCompleted, b.JiraIssuesCompleted),
		JiraStoryPoints:     a.JiraStoryPoints + b.JiraStoryPoints,
		JiraVelocity:        combinedVelocity.Value(),
		JiraComments:        intSum.Combine(a.JiraComments, b.JiraComments),
		JiraTransitions:     intSum.Combine(a.JiraTransitions, b.JiraTransitions),

		// Zoom metrics (sum)
		ZoomMeetings:   intSum.Combine(a.ZoomMeetings, b.ZoomMeetings),
		ZoomMinutes:    intSum.Combine(a.ZoomMinutes, b.ZoomMinutes),
		ZoomWebinars:   intSum.Combine(a.ZoomWebinars, b.ZoomWebinars),
		ZoomRecordings: intSum.Combine(a.ZoomRecordings, b.ZoomRecordings),

		// Derived metrics
		ActiveHours:        activeHoursSum,
		PeakHour:           peakHour,
		CollaborationScore: combinedCollab.Value(),

		// Metadata
		DataPoints: intSum.Combine(a.DataPoints, b.DataPoints),
	}
}

// Verify StructuredMetricsMonoid implements Monoid interface
var _ monoid.Monoid[analytics.StructuredMetrics] = StructuredMetricsMonoid{}

// TotalActivity helper for determining peak hour
func (sm analytics.StructuredMetrics) TotalActivity() int {
	return sm.SlackMessages + sm.GitHubCommits + sm.JiraIssuesCompleted + sm.ZoomMeetings
}
