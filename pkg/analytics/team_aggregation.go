package analytics

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/effect"
	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purekernels/pkg/result"
	"github.com/vinodhalaharvi/purepulse/db/query"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// TYPE ALIASES FOR TEAM AGGREGATION
// ============================================================================

type TotalUserCount int
type TotalWins int
type TotalBlockers int
type AverageProductivity float64
type AverageCollaboration float64

// ============================================================================
// DAILY TEAM AGGREGATION
// ============================================================================

type DailyTeamAggregation struct {
	TeamID            types.TeamID
	Date              time.Time
	UserCount         TotalUserCount
	TotalEventCount   int
	PlatformBreakdown map[types.Platform]int
	TopActiveUsers    []types.UserID
	GeneratedAt       time.Time
}

// AggregateUserDailyActivityToTeam aggregates individual user daily activity
func AggregateUserDailyActivityToTeam(
	teamID types.TeamID,
	date time.Time,
	userActivities map[types.UserID][]query.DailyActivityAgg,
) DailyTeamAggregation {

	aggregation := DailyTeamAggregation{
		TeamID:            teamID,
		Date:              date,
		UserCount:         TotalUserCount(len(userActivities)),
		PlatformBreakdown: make(map[types.Platform]int),
		TopActiveUsers:    []types.UserID{},
		GeneratedAt:       time.Now(),
	}

	// Track events per user for ranking
	type userActivity struct {
		userID     types.UserID
		eventCount int
	}

	userEvents := make([]userActivity, 0)

	for userID, activities := range userActivities {
		userEventCount := 0

		for _, activity := range activities {
			aggregation.TotalEventCount += activity.EventCount
			aggregation.PlatformBreakdown[activity.Source] += activity.EventCount
			userEventCount += activity.EventCount
		}

		userEvents = append(userEvents, userActivity{userID, userEventCount})
	}

	// Sort users by event count and get top 3
	sort.Slice(userEvents, func(i, j int) bool {
		return userEvents[i].eventCount > userEvents[j].eventCount
	})

	for i, ua := range userEvents {
		if i >= 3 {
			break
		}
		aggregation.TopActiveUsers = append(aggregation.TopActiveUsers, ua.userID)
	}

	return aggregation
}

// ============================================================================
// WEEKLY TEAM AGGREGATION
// ============================================================================

type WeeklyTeamAggregationMetrics struct {
	TeamID                    types.TeamID
	WeekStart                 time.Time
	WeekEnd                   time.Time
	TotalUserCount            TotalUserCount
	TotalEventCount           int
	AveragePlatforms          float64
	TopCollaborators          []types.UserID
	TotalWinsReported         TotalWins
	TotalBlockersReported     TotalBlockers
	AverageProductivityScore  AverageProductivity
	AverageCollaborationScore AverageCollaboration
	GeneratedAt               time.Time
}

// AggregateUserReportsToTeamMetrics aggregates individual user reports to team level
func AggregateUserReportsToTeamMetrics(
	teamID types.TeamID,
	week types.TimeRange,
	userReports map[types.UserID]UserWeeklyReport,
) WeeklyTeamAggregationMetrics {

	metrics := WeeklyTeamAggregationMetrics{
		TeamID:           teamID,
		WeekStart:        week.Start,
		WeekEnd:          week.End,
		TotalUserCount:   TotalUserCount(len(userReports)),
		TopCollaborators: []types.UserID{},
		GeneratedAt:      time.Now(),
	}

	platformSet := make(map[types.Platform]bool)
	var productivityScores []float64
	var collaborationScores []float64

	// Aggregate metrics across all users
	for userID, report := range userReports {
		metrics.TotalWinsReported += TotalWins(len(report.Wins))
		metrics.TotalBlockersReported += TotalBlockers(len(report.Blocked))

		// Collect scores for averaging
		productivityScores = append(productivityScores, float64(0))   // Would come from user metrics
		collaborationScores = append(collaborationScores, float64(0)) // Would come from user metrics

		// Track active platforms (simplified)
		_ = userID
	}

	// Calculate averages
	if len(productivityScores) > 0 {
		sum := 0.0
		for _, score := range productivityScores {
			sum += score
		}
		metrics.AverageProductivityScore = AverageProductivity(sum / float64(len(productivityScores)))
	}

	if len(collaborationScores) > 0 {
		sum := 0.0
		for _, score := range collaborationScores {
			sum += score
		}
		metrics.AverageCollaborationScore = AverageCollaboration(sum / float64(len(collaborationScores)))
	}

	if len(platformSet) > 0 {
		metrics.AveragePlatforms = float64(len(platformSet))
	}

	return metrics
}

// ============================================================================
// MONOID INSTANCES FOR TEAM AGGREGATION
// ============================================================================

type DailyTeamAggregationMonoid struct{}

var _ monoid.Monoid[DailyTeamAggregation] = DailyTeamAggregationMonoid{}

func (DailyTeamAggregationMonoid) Empty() DailyTeamAggregation {
	return DailyTeamAggregation{
		PlatformBreakdown: make(map[types.Platform]int),
		TopActiveUsers:    []types.UserID{},
		GeneratedAt:       time.Now(),
	}
}

func (DailyTeamAggregationMonoid) Combine(a, b DailyTeamAggregation) DailyTeamAggregation {
	combined := DailyTeamAggregation{
		TeamID:            a.TeamID,
		Date:              a.Date,
		TotalEventCount:   a.TotalEventCount + b.TotalEventCount,
		PlatformBreakdown: make(map[types.Platform]int),
		TopActiveUsers:    []types.UserID{},
		GeneratedAt:       time.Now(),
	}

	// Combine platform breakdowns
	for platform, count := range a.PlatformBreakdown {
		combined.PlatformBreakdown[platform] += count
	}
	for platform, count := range b.PlatformBreakdown {
		combined.PlatformBreakdown[platform] += count
	}

	// Combine user lists (dedup)
	userMap := make(map[types.UserID]bool)
	for _, user := range a.TopActiveUsers {
		userMap[user] = true
	}
	for _, user := range b.TopActiveUsers {
		userMap[user] = true
	}
	for user := range userMap {
		combined.TopActiveUsers = append(combined.TopActiveUsers, user)
	}

	return combined
}

type WeeklyTeamAggregationMetricsMonoid struct{}

var _ monoid.Monoid[WeeklyTeamAggregationMetrics] = WeeklyTeamAggregationMetricsMonoid{}

func (WeeklyTeamAggregationMetricsMonoid) Empty() WeeklyTeamAggregationMetrics {
	return WeeklyTeamAggregationMetrics{
		TopCollaborators: []types.UserID{},
		GeneratedAt:      time.Now(),
	}
}

func (WeeklyTeamAggregationMetricsMonoid) Combine(
	a, b WeeklyTeamAggregationMetrics,
) WeeklyTeamAggregationMetrics {

	combined := WeeklyTeamAggregationMetrics{
		TeamID:                    a.TeamID,
		WeekStart:                 a.WeekStart,
		WeekEnd:                   a.WeekEnd,
		TotalUserCount:            a.TotalUserCount + b.TotalUserCount,
		TotalEventCount:           a.TotalEventCount + b.TotalEventCount,
		TotalWinsReported:         a.TotalWinsReported + b.TotalWinsReported,
		TotalBlockersReported:     a.TotalBlockersReported + b.TotalBlockersReported,
		AverageProductivityScore:  (a.AverageProductivityScore + b.AverageProductivityScore) / 2,
		AverageCollaborationScore: (a.AverageCollaborationScore + b.AverageCollaborationScore) / 2,
		GeneratedAt:               time.Now(),
	}

	// Combine collaborators
	collabMap := make(map[types.UserID]bool)
	for _, user := range a.TopCollaborators {
		collabMap[user] = true
	}
	for _, user := range b.TopCollaborators {
		collabMap[user] = true
	}
	for user := range collabMap {
		combined.TopCollaborators = append(combined.TopCollaborators, user)
	}

	return combined
}

// ============================================================================
// IMPURE OPERATIONS (Wrapped in effect.Writer)
// ============================================================================

// FetchAndAggregateDailyTeamActivity fetches all users' daily activity and aggregates
func FetchAndAggregateDailyTeamActivity(
	ctx context.Context,
	conn *interface{}, // Would be db.Connection
	teamID types.TeamID,
	userIDs []types.UserID,
	date time.Time,
) effect.Writer[[]string, result.Result[DailyTeamAggregation]] {

	logs := []string{fmt.Sprintf("fetch_daily_team_started: team=%s, users=%d", teamID, len(userIDs))}

	// Placeholder - would fetch from database
	_ = conn
	_ = ctx

	aggregation := DailyTeamAggregation{
		TeamID:            teamID,
		Date:              date,
		UserCount:         TotalUserCount(len(userIDs)),
		PlatformBreakdown: make(map[types.Platform]int),
		TopActiveUsers:    userIDs,
		GeneratedAt:       time.Now(),
	}

	logs = append(logs, fmt.Sprintf("fetch_daily_team_succeeded: users=%d", len(userIDs)))
	return effect.NewWriter(result.Ok(aggregation), logs)
}

// FetchAndAggregateWeeklyTeamReports aggregates individual user reports to team metrics
func FetchAndAggregateWeeklyTeamReports(
	ctx context.Context,
	teamID types.TeamID,
	week types.TimeRange,
	userReports map[types.UserID]UserWeeklyReport,
) effect.Writer[[]string, result.Result[WeeklyTeamAggregationMetrics]] {

	logs := []string{fmt.Sprintf("aggregate_weekly_team_started: team=%s, users=%d", teamID, len(userReports))}

	_ = ctx // Context for potential future async operations

	metrics := AggregateUserReportsToTeamMetrics(teamID, week, userReports)

	logs = append(logs, fmt.Sprintf("aggregate_weekly_team_succeeded: wins=%d, blockers=%d",
		metrics.TotalWinsReported, metrics.TotalBlockersReported))

	return effect.NewWriter(result.Ok(metrics), logs)
}
