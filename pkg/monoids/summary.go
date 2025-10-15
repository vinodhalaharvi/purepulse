package monoids

import (
	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// AI SUMMARY MONOID
// ============================================================================

// AISummaryMonoid combines AI-generated summaries
type AISummaryMonoid struct{}

// Empty returns the identity element for AISummary
func (AISummaryMonoid) Empty() analytics.AISummary {
	return analytics.AISummary{
		Highlights:      []string{},
		Recommendations: []string{},
	}
}

// Combine merges two AISummary instances
func (m AISummaryMonoid) Combine(a, b analytics.AISummary) analytics.AISummary {
	// Use list monoid for string slices
	stringListMonoid := monoid.NewListMonoid[string]()

	// Use sum monoid for tokens
	intSum := monoid.NewSumMonoid[int]()

	// Concatenate overviews with separator
	overview := ""
	if a.Overview != "" && b.Overview != "" {
		overview = a.Overview + "\n\n" + b.Overview
	} else if a.Overview != "" {
		overview = a.Overview
	} else {
		overview = b.Overview
	}

	// Use max for latency (parallel calls)
	latency := a.Metadata.LatencyMS
	if b.Metadata.LatencyMS > latency {
		latency = b.Metadata.LatencyMS
	}

	return analytics.AISummary{
		Overview:        overview,
		Highlights:      stringListMonoid.Combine(a.Highlights, b.Highlights),
		Recommendations: stringListMonoid.Combine(a.Recommendations, b.Recommendations),
		Metadata: analytics.LLMMetadata{
			Model:        coalesceString(a.Metadata.Model, b.Metadata.Model),
			Tokens:       intSum.Combine(a.Metadata.Tokens, b.Metadata.Tokens),
			LatencyMS:    latency,
			CachedPrompt: a.Metadata.CachedPrompt || b.Metadata.CachedPrompt,
			Cost:         a.Metadata.Cost + b.Metadata.Cost,
		},
	}
}

// Verify AISummaryMonoid implements Monoid interface
var _ monoid.Monoid[analytics.AISummary] = AISummaryMonoid{}

// ============================================================================
// USER SUMMARY MONOID
// ============================================================================

// UserSummaryMonoid combines complete user summaries
type UserSummaryMonoid struct{}

// Empty returns the identity element for UserSummary
func (UserSummaryMonoid) Empty() analytics.UserSummary {
	return analytics.UserSummary{
		Activity:     UserActivityMonoid{}.Empty(),
		Metrics:      StructuredMetricsMonoid{}.Empty(),
		Correlations: []analytics.Correlation{},
		AISummary:    AISummaryMonoid{}.Empty(),
		AuditLog:     []string{},
	}
}

// Combine merges two UserSummary instances
func (m UserSummaryMonoid) Combine(a, b analytics.UserSummary) analytics.UserSummary {
	activityMonoid := UserActivityMonoid{}
	metricsMonoid := StructuredMetricsMonoid{}
	aiSummaryMonoid := AISummaryMonoid{}
	correlationListMonoid := monoid.NewListMonoid[analytics.Correlation]()
	auditLogMonoid := monoid.NewListMonoid[string]()

	// Use latest GeneratedAt
	generatedAt := a.GeneratedAt
	if b.GeneratedAt.After(a.GeneratedAt) {
		generatedAt = b.GeneratedAt
	}

	return analytics.UserSummary{
		UserID:       coalesceUserID(a.UserID, b.UserID),
		TimeRange:    coalesceTimeRange(a.TimeRange, b.TimeRange),
		Activity:     activityMonoid.Combine(a.Activity, b.Activity),
		Metrics:      metricsMonoid.Combine(a.Metrics, b.Metrics),
		Correlations: correlationListMonoid.Combine(a.Correlations, b.Correlations),
		AISummary:    aiSummaryMonoid.Combine(a.AISummary, b.AISummary),
		GeneratedAt:  generatedAt,
		AuditLog:     auditLogMonoid.Combine(a.AuditLog, b.AuditLog),
		Version:      coalesceString(a.Version, b.Version),
	}
}

// Verify UserSummaryMonoid implements Monoid interface
var _ monoid.Monoid[analytics.UserSummary] = UserSummaryMonoid{}

// ============================================================================
// TEAM SUMMARY MONOID
// ============================================================================

// TeamSummaryMonoid combines team summaries
type TeamSummaryMonoid struct{}

// Empty returns the identity element for TeamSummary
func (TeamSummaryMonoid) Empty() analytics.TeamSummary {
	return analytics.TeamSummary{
		Members:    []analytics.UserSummary{},
		Aggregated: AggregatedTeamMetricsMonoid{}.Empty(),
	}
}

// Combine merges two TeamSummary instances
func (m TeamSummaryMonoid) Combine(a, b analytics.TeamSummary) analytics.TeamSummary {
	memberListMonoid := monoid.NewListMonoid[analytics.UserSummary]()
	aggregatedMonoid := AggregatedTeamMetricsMonoid{}

	// Use latest GeneratedAt
	generatedAt := a.GeneratedAt
	if b.GeneratedAt.After(a.GeneratedAt) {
		generatedAt = b.GeneratedAt
	}

	return analytics.TeamSummary{
		TeamID:      coalesceTeamID(a.TeamID, b.TeamID),
		TimeRange:   coalesceTimeRange(a.TimeRange, b.TimeRange),
		Members:     memberListMonoid.Combine(a.Members, b.Members),
		Aggregated:  aggregatedMonoid.Combine(a.Aggregated, b.Aggregated),
		GeneratedAt: generatedAt,
	}
}

// Verify TeamSummaryMonoid implements Monoid interface
var _ monoid.Monoid[analytics.TeamSummary] = TeamSummaryMonoid{}

// ============================================================================
// AGGREGATED TEAM METRICS MONOID
// ============================================================================

// AggregatedTeamMetricsMonoid combines team-level metrics
type AggregatedTeamMetricsMonoid struct{}

// Empty returns the identity element for AggregatedTeamMetrics
func (AggregatedTeamMetricsMonoid) Empty() analytics.AggregatedTeamMetrics {
	return analytics.AggregatedTeamMetrics{
		TopContributors: []analytics.TopContributor{},
	}
}

// Combine merges two AggregatedTeamMetrics instances
func (m AggregatedTeamMetricsMonoid) Combine(a, b analytics.AggregatedTeamMetrics) analytics.AggregatedTeamMetrics {
	intSum := monoid.NewSumMonoid[int]()

	// Combine averages using weighted average
	teamVelocity := weightedAverage(
		a.TeamVelocity, float64(a.TotalMembers),
		b.TeamVelocity, float64(b.TotalMembers),
	)

	collaborationScore := weightedAverage(
		a.CollaborationScore, float64(a.TotalMembers),
		b.CollaborationScore, float64(b.TotalMembers),
	)

	codeQualityAvg := weightedAverage(
		a.CodeQualityAvg, float64(a.TotalMembers),
		b.CodeQualityAvg, float64(b.TotalMembers),
	)

	// Merge top contributors (keeping top N)
	contributorListMonoid := monoid.NewListMonoid[analytics.TopContributor]()
	mergedContributors := contributorListMonoid.Combine(a.TopContributors, b.TopContributors)

	return analytics.AggregatedTeamMetrics{
		TotalMembers:  intSum.Combine(a.TotalMembers, b.TotalMembers),
		ActiveMembers: intSum.Combine(a.ActiveMembers, b.ActiveMembers),

		TotalMessages: intSum.Combine(a.TotalMessages, b.TotalMessages),
		TotalCommits:  intSum.Combine(a.TotalCommits, b.TotalCommits),
		TotalPRs:      intSum.Combine(a.TotalPRs, b.TotalPRs),
		TotalIssues:   intSum.Combine(a.TotalIssues, b.TotalIssues),
		TotalMeetings: intSum.Combine(a.TotalMeetings, b.TotalMeetings),

		TeamVelocity:       teamVelocity,
		CollaborationScore: collaborationScore,
		CodeQualityAvg:     codeQualityAvg,

		TopContributors: mergedContributors,
	}
}

// Verify AggregatedTeamMetricsMonoid implements Monoid interface
var _ monoid.Monoid[analytics.AggregatedTeamMetrics] = AggregatedTeamMetricsMonoid{}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// coalesceTeamID returns first non-empty TeamID
func coalesceTeamID(a, b types.TeamID) types.TeamID {
	if a != "" {
		return a
	}
	return b
}
