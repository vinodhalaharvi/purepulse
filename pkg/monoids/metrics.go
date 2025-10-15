package monoids

import (
	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
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

	// For averages, use weighted average formula manually
	// weighted avg = (val1*weight1 + val2*weight2) / (weight1 + weight2)

	// Code quality weighted by commit count
	codeQuality := weightedAverage(
		a.CodeQuality, float64(a.GitHubCommits),
		b.CodeQuality, float64(b.GitHubCommits),
	)

	// Jira velocity weighted by data points
	jiraVelocity := weightedAverage(
		a.JiraVelocity, float64(a.DataPoints),
		b.JiraVelocity, float64(b.DataPoints),
	)

	// Collaboration score weighted by data points
	collaborationScore := weightedAverage(
		a.CollaborationScore, float64(a.DataPoints),
		b.CollaborationScore, float64(b.DataPoints),
	)

	// Active hours as sum
	activeHoursSum := a.ActiveHours + b.ActiveHours

	// Peak hour: use the one with more activity
	peakHour := a.PeakHour
	totalActivityA := totalActivityCount(a)
	totalActivityB := totalActivityCount(b)
	if totalActivityB > totalActivityA {
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
		CodeQuality:        codeQuality,

		// Jira metrics (sum)
		JiraIssuesCreated:   intSum.Combine(a.JiraIssuesCreated, b.JiraIssuesCreated),
		JiraIssuesCompleted: intSum.Combine(a.JiraIssuesCompleted, b.JiraIssuesCompleted),
		JiraStoryPoints:     a.JiraStoryPoints + b.JiraStoryPoints,
		JiraVelocity:        jiraVelocity,
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
		CollaborationScore: collaborationScore,

		// Metadata
		DataPoints: intSum.Combine(a.DataPoints, b.DataPoints),
	}
}

// Verify StructuredMetricsMonoid implements Monoid interface
var _ monoid.Monoid[analytics.StructuredMetrics] = StructuredMetricsMonoid{}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// weightedAverage computes weighted average of two values
func weightedAverage(val1, weight1, val2, weight2 float64) float64 {
	totalWeight := weight1 + weight2
	if totalWeight == 0 {
		return 0
	}
	return (val1*weight1 + val2*weight2) / totalWeight
}

// totalActivityCount returns total activity count for peak hour calculation
func totalActivityCount(sm analytics.StructuredMetrics) int {
	return sm.SlackMessages + sm.GitHubCommits + sm.JiraIssuesCompleted + sm.ZoomMeetings
}
