// pkg/analytics/aggregation.go
package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/effect"
	"github.com/vinodhalaharvi/purekernels/pkg/result"
	"github.com/vinodhalaharvi/purepulse/db/query"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// QUERY BUILDERS USING MONOIDS (Pure)
// ============================================================================

func BuildDailyActivityQuery(
	userID types.UserID,
	dateRange types.TimeRange,
) query.Query[query.DailyActivityAgg] {
	userFilter := query.Where("user_id", string(userID))
	dateFilter := query.WhereBetween("date", dateRange.Start, dateRange.End)

	return query.Table[query.DailyActivityAgg]("daily_user_activity").
		Filter(userFilter.Combine(dateFilter)).
		Sort(query.Desc("date"))
}

func BuildWeeklyTeamActivityQuery(
	teamID types.TeamID,
	weekStart time.Time,
) query.Query[query.TeamActivityRow] {
	weekEnd := weekStart.Add(7 * 24 * time.Hour)

	teamFilter := query.Where("team_id", string(teamID))
	dateFilter := query.WhereBetween("week_start", weekStart, weekEnd)

	return query.Table[query.TeamActivityRow]("weekly_team_activity").
		Filter(teamFilter.Combine(dateFilter)).
		Sort(query.Desc("week_start"))
}

func BuildCorrelationSummaryQuery(
	userID types.UserID,
) query.Query[query.CorrelationRow] {
	userFilter := query.Where("user_id", string(userID))

	return query.Table[query.CorrelationRow]("user_correlation_summary").
		Filter(userFilter).
		Sort(query.Desc("avg_confidence"))
}

// ============================================================================
// AGGREGATION FROM VIEW RESULTS (Pure)
// ============================================================================

func AggregateUserMetricsFromViews(
	userID types.UserID,
	week types.TimeRange,
	dailyActivityRows []query.DailyActivityAgg,
	correlationRows []query.CorrelationRow,
) WeeklyUserMetrics {

	metrics := WeeklyUserMetrics{
		UserID:      userID,
		Week:        week,
		GitHub:      aggregateGitHubMetrics(dailyActivityRows),
		Jira:        aggregateJiraMetrics(dailyActivityRows),
		Slack:       aggregateSlackMetrics(dailyActivityRows),
		Meetings:    aggregateMeetingMetrics(dailyActivityRows),
		GeneratedAt: time.Now(),
	}

	metrics.Patterns = ActivityPatterns{
		PeakActivityHours:  DetectPeakHours(dailyActivityRows),
		CollaborationScore: ComputeCollaborationScore(metrics),
		ProductivityScore:  ComputeProductivityScore(metrics),
	}

	return metrics
}

// ============================================================================
// TEAM AGGREGATION
// ============================================================================

func AggregateTeamMetricsFromViews(
	teamID types.TeamID,
	week types.TimeRange,
	teamActivityRows []query.TeamActivityRow,
	userMetrics []WeeklyUserMetrics,
) WeeklyTeamMetrics {

	teamMetrics := WeeklyTeamMetrics{
		TeamID:           teamID,
		Week:             week,
		Members:          userMetrics,
		Velocity:         TeamVelocity(len(userMetrics) * 5),
		TopCollaborators: extractTopCollaborators(userMetrics),
		KeyBlockers:      extractKeyBlockers(userMetrics),
		MentorshipPairs:  detectTeamMentorship(userMetrics),
		GeneratedAt:      time.Now(),
	}

	return teamMetrics
}

func extractTopCollaborators(userMetrics []WeeklyUserMetrics) []types.UserID {
	type userScore struct {
		userID types.UserID
		score  CollaborationScore
	}

	scores := make([]userScore, len(userMetrics))
	for i, m := range userMetrics {
		scores[i] = userScore{m.UserID, m.Patterns.CollaborationScore}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	result := make([]types.UserID, len(scores))
	for i, s := range scores {
		result[i] = s.userID
	}

	return result
}

func extractKeyBlockers(userMetrics []WeeklyUserMetrics) []string {
	blockers := make([]string, 0)

	for _, m := range userMetrics {
		if m.Jira.TicketsBlocked > 0 {
			blockers = append(blockers, fmt.Sprintf("%s: %d blocked tickets", m.UserID, m.Jira.TicketsBlocked))
		}
	}

	return blockers
}

func detectTeamMentorship(userMetrics []WeeklyUserMetrics) []MentorshipRelation {
	mentorship := make([]MentorshipRelation, 0)

	for i, mentor := range userMetrics {
		for j, mentee := range userMetrics {
			if i != j && mentor.Patterns.ProductivityScore > mentee.Patterns.ProductivityScore {
				if mentor.Slack.HelpQuestionsAnswered > 0 {
					mentorship = append(mentorship, MentorshipRelation{
						Mentor:               mentor.UserID,
						Mentee:               mentee.UserID,
						PairProgrammingHours: mentor.Slack.PairProgrammingHours,
						FocusArea:            "general",
					})
				}
			}
		}
	}

	return mentorship
}

// ============================================================================
// PATTERN DETECTION (Pure)
// ============================================================================

func DetectContextSwitches(dailyActivityRows []query.DailyActivityAgg) int {
	platforms := make(map[types.Platform]bool)

	for _, row := range dailyActivityRows {
		platforms[row.Source] = true
	}

	return len(platforms)
}

func ComputeCollaborationScore(userMetrics WeeklyUserMetrics) CollaborationScore {
	score := CollaborationScore(0.0)

	if userMetrics.Slack.Messages > 0 {
		score += CollaborationScore(0.4 * float64(userMetrics.Slack.HelpQuestionsAnswered) / float64(userMetrics.Slack.Messages+1))
	}

	if userMetrics.Slack.PairProgrammingHours > 0 {
		score += CollaborationScore(0.4)
	}

	score += CollaborationScore(0.2 * float64(userMetrics.GitHub.PRsReviewed) / 10.0)

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func ComputeProductivityScore(userMetrics WeeklyUserMetrics) ProductivityScore {
	score := ProductivityScore(0.0)

	commitScore := float64(userMetrics.GitHub.Commits) / 20.0
	if commitScore > 0.4 {
		commitScore = 0.4
	}
	score += ProductivityScore(commitScore)

	jiraScore := float64(userMetrics.Jira.TicketsCompleted) / 10.0
	if jiraScore > 0.4 {
		jiraScore = 0.4
	}
	score += ProductivityScore(jiraScore)

	deepWorkScore := float64(userMetrics.Meetings.DeepWorkHours) / 40.0
	if deepWorkScore > 0.2 {
		deepWorkScore = 0.2
	}
	score += ProductivityScore(deepWorkScore)

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// ============================================================================
// DATABASE QUERIES (Impure - wrapped in effect.Writer)
// ============================================================================

func FetchWeeklyMetricsForUser(
	ctx context.Context,
	conn *sql.DB,
	userID types.UserID,
	week types.TimeRange,
) effect.Writer[[]string, result.Result[WeeklyUserMetrics]] {

	logs := []string{fmt.Sprintf("fetch_weekly_metrics_started: user=%s", userID)}

	dailyQuery := BuildDailyActivityQuery(userID, week)
	sql, params := dailyQuery.Build()

	dailyRows, err := conn.QueryContext(ctx, sql, params...)
	if err != nil {
		logs = append(logs, fmt.Sprintf("fetch_failed: %v", err))
		return effect.NewWriter(result.Err[WeeklyUserMetrics](err), logs)
	}
	defer dailyRows.Close()

	var dailyActivityRows []query.DailyActivityAgg
	for dailyRows.Next() {
		row, err := query.ScanDailyActivityAgg(dailyRows)
		if err != nil {
			logs = append(logs, fmt.Sprintf("scan_failed: %v", err))
			continue
		}
		dailyActivityRows = append(dailyActivityRows, row)
	}

	logs = append(logs, fmt.Sprintf("fetched %d daily activity rows", len(dailyActivityRows)))

	corrQuery := BuildCorrelationSummaryQuery(userID)
	corrSQL, corrParams := corrQuery.Build()

	corrRows, err := conn.QueryContext(ctx, corrSQL, corrParams...)
	if err != nil {
		logs = append(logs, fmt.Sprintf("correlation_fetch_failed: %v", err))
		return effect.NewWriter(result.Err[WeeklyUserMetrics](err), logs)
	}
	defer corrRows.Close()

	var correlationRows []query.CorrelationRow
	for corrRows.Next() {
		row, err := query.ScanCorrelationRow(corrRows)
		if err != nil {
			logs = append(logs, fmt.Sprintf("correlation_scan_failed: %v", err))
			continue
		}
		correlationRows = append(correlationRows, row)
	}

	logs = append(logs, fmt.Sprintf("fetched %d correlation rows", len(correlationRows)))

	metrics := AggregateUserMetricsFromViews(userID, week, dailyActivityRows, correlationRows)

	logs = append(logs, fmt.Sprintf("fetch_weekly_metrics_succeeded: collaboration=%.2f, productivity=%.2f",
		metrics.Patterns.CollaborationScore, metrics.Patterns.ProductivityScore))

	return effect.NewWriter(result.Ok(metrics), logs)
}

func FetchWeeklyMetricsForTeam(
	ctx context.Context,
	conn *sql.DB,
	teamID types.TeamID,
	week types.TimeRange,
) effect.Writer[[]string, result.Result[WeeklyTeamMetrics]] {

	logs := []string{fmt.Sprintf("fetch_team_metrics_started: team=%s", teamID)}

	teamQuery := BuildWeeklyTeamActivityQuery(teamID, week.Start)
	sql, params := teamQuery.Build()

	teamRows, err := conn.QueryContext(ctx, sql, params...)
	if err != nil {
		logs = append(logs, fmt.Sprintf("fetch_failed: %v", err))
		return effect.NewWriter(result.Err[WeeklyTeamMetrics](err), logs)
	}
	defer teamRows.Close()

	var teamActivityRows []query.TeamActivityRow
	for teamRows.Next() {
		row, err := query.ScanTeamActivityRow(teamRows)
		if err != nil {
			logs = append(logs, fmt.Sprintf("scan_failed: %v", err))
			continue
		}
		teamActivityRows = append(teamActivityRows, row)
	}

	logs = append(logs, fmt.Sprintf("fetched %d team activity rows", len(teamActivityRows)))

	metrics := AggregateTeamMetricsFromViews(teamID, week, teamActivityRows, []WeeklyUserMetrics{})

	logs = append(logs, fmt.Sprintf("fetch_team_metrics_succeeded: velocity=%d, blockers=%d",
		metrics.Velocity, metrics.BlockerCount))

	return effect.NewWriter(result.Ok(metrics), logs)
}

func SaveWeeklyMetrics(
	ctx context.Context,
	conn *sql.DB,
	metrics WeeklyTeamMetrics,
) effect.Writer[[]string, result.Result[string]] {

	logs := []string{fmt.Sprintf("save_metrics_started: team=%s", metrics.TeamID)}

	stmt := `
        INSERT INTO team_weekly_metrics 
        (team_id, week_start, week_end, total_velocity, velocity_trend, blocker_count, team_collaboration_score)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (team_id, week_start, week_end) DO UPDATE SET
            total_velocity = $4,
            velocity_trend = $5,
            blocker_count = $6,
            team_collaboration_score = $7,
            updated_at = NOW()
    `

	_, err := conn.ExecContext(ctx, stmt,
		metrics.TeamID,
		metrics.Week.Start,
		metrics.Week.End,
		metrics.Velocity,
		metrics.VelocityTrend,
		metrics.BlockerCount,
		metrics.CollaborationScore,
	)

	if err != nil {
		logs = append(logs, fmt.Sprintf("save_failed: %v", err))
		return effect.NewWriter(result.Err[string](err), logs)
	}

	logs = append(logs, "save_metrics_succeeded")
	return effect.NewWriter(result.Ok(fmt.Sprintf("Saved metrics for %s", metrics.TeamID)), logs)
}
