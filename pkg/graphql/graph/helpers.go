package graph

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
	"github.com/vinodhalaharvi/purepulse/pkg/graphql/graph/model"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// SimpleDailyActivity is a simplified version for now
type SimpleDailyActivity struct {
	UserID     string
	Date       time.Time
	Source     string
	EventCount int
	Channels   pq.StringArray
	TotalSize  int
}

// User Weekly Report Helpers

func (r *queryResolver) fetchExistingReport(ctx context.Context, userID string, weekStart time.Time) (*model.UserWeeklyReport, error) {
	query := `
        SELECT 
            wins,
            in_progress,
            blocked,
            notes,
            generated_at
        FROM weekly_reports
        WHERE user_id = $1 
        AND week_start = $2
        ORDER BY generated_at DESC
        LIMIT 1
    `

	var winsJSON, progressJSON, blockedJSON []byte
	var notes sql.NullString
	var generatedAt time.Time

	err := r.DB.QueryRowContext(ctx, query, userID, weekStart).Scan(
		&winsJSON,
		&progressJSON,
		&blockedJSON,
		&notes,
		&generatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No existing report
	}
	if err != nil {
		return nil, err
	}

	// Parse the JSON fields
	var wins []analytics.WinItem
	var inProgress []analytics.ProgressItem
	var blocked []analytics.BlockedItem

	json.Unmarshal(winsJSON, &wins)
	json.Unmarshal(progressJSON, &inProgress)
	json.Unmarshal(blockedJSON, &blocked)

	// Convert to GraphQL model
	report := &model.UserWeeklyReport{
		UserID:      userID,
		Wins:        make([]*model.Win, 0, len(wins)),
		InProgress:  make([]*model.InProgress, 0, len(inProgress)),
		Blocked:     make([]*model.Blocked, 0, len(blocked)),
		GeneratedAt: generatedAt.Format(time.RFC3339),
	}

	for _, w := range wins {
		report.Wins = append(report.Wins, &model.Win{
			Title:       w.Title,
			Description: w.Description,
			Impact:      w.Impact,
		})
	}

	for _, p := range inProgress {
		var blocker, dueDate *string
		if p.Blocker != "" {
			blocker = &p.Blocker
		}
		if !p.DueDate.IsZero() {
			dueDateStr := p.DueDate.Format("2006-01-02")
			dueDate = &dueDateStr
		}
		report.InProgress = append(report.InProgress, &model.InProgress{
			Title:           p.Title,
			PercentComplete: int32(p.PercentComplete),
			Blocker:         blocker,
			DueDate:         dueDate,
		})
	}

	for _, b := range blocked {
		report.Blocked = append(report.Blocked, &model.Blocked{
			Title:           b.Title,
			BlockedBy:       b.BlockedBy,
			DurationHours:   int32(b.DurationHours),
			ImpactLevel:     string(b.ImpactLevel),
			SuggestedAction: b.SuggestedAction,
		})
	}

	if notes.Valid {
		report.Notes = &notes.String
	}

	return report, nil
}

func (r *queryResolver) fetchDailyActivity(ctx context.Context, userID string, timeRange types.TimeRange) ([]SimpleDailyActivity, error) {
	query := `
        SELECT 
            user_id, 
            date, 
            source, 
            event_count,
            channels,
            total_size
        FROM daily_user_activity
        WHERE user_id = $1 AND date >= $2 AND date <= $3
        ORDER BY date DESC
        LIMIT 50
    `

	rows, err := r.DB.QueryContext(ctx, query, userID, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []SimpleDailyActivity
	for rows.Next() {
		var a SimpleDailyActivity
		err := rows.Scan(
			&a.UserID,
			&a.Date,
			&a.Source,
			&a.EventCount,
			&a.Channels,
			&a.TotalSize,
		)
		if err != nil {
			log.Printf("Warning: failed to scan row: %v", err)
			continue
		}
		activities = append(activities, a)
	}

	return activities, nil
}

func (r *queryResolver) buildBasicReport(userID string, dailyActivity []SimpleDailyActivity) *model.UserWeeklyReport {
	report := &model.UserWeeklyReport{
		UserID:      userID,
		Wins:        []*model.Win{},
		InProgress:  []*model.InProgress{},
		Blocked:     []*model.Blocked{},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	// Basic analysis without LLM
	totalEvents := 0
	platforms := make(map[string]int)

	for _, day := range dailyActivity {
		totalEvents += day.EventCount
		platforms[day.Source]++
	}

	// Generate basic wins
	if totalEvents > 50 {
		report.Wins = append(report.Wins, &model.Win{
			Title:       fmt.Sprintf("High activity: %d events", totalEvents),
			Description: "Maintained strong engagement across platforms",
			Impact:      "Medium",
		})
	}

	for platform, count := range platforms {
		if count > 5 {
			report.Wins = append(report.Wins, &model.Win{
				Title:       fmt.Sprintf("Active on %s", platform),
				Description: fmt.Sprintf("%d days of activity", count),
				Impact:      "Low",
			})
		}
	}

	// Add a sample in-progress item
	report.InProgress = append(report.InProgress, &model.InProgress{
		Title:           "Ongoing tasks",
		PercentComplete: 50,
		Blocker:         nil,
		DueDate:         nil,
	})

	notes := fmt.Sprintf("Week summary: %d total events across %d platforms", totalEvents, len(platforms))
	report.Notes = &notes

	return report
}

// Team Weekly Report Helpers

func (r *queryResolver) fetchExistingTeamReport(ctx context.Context, teamID string, weekStart time.Time) (*model.TeamWeeklyReport, error) {
	query := `
		SELECT 
			executive_summary,
			velocity_analysis,
			collaboration_notes,
			team_blockers,
			recommendations,
			generated_at
		FROM team_weekly_reports
		WHERE team_id = $1 AND week_start::date = $2::date
		ORDER BY generated_at DESC
		LIMIT 1
	`

	var executiveSummary, velocityAnalysis, collaborationNotes string
	var blockersJSON, recommendationsJSON []byte
	var generatedAt time.Time

	err := r.DB.QueryRowContext(ctx, query, teamID, weekStart).Scan(
		&executiveSummary,
		&velocityAnalysis,
		&collaborationNotes,
		&blockersJSON,
		&recommendationsJSON,
		&generatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var blockers []analytics.TeamBlockerItem
	var recommendations []string
	json.Unmarshal(blockersJSON, &blockers)
	json.Unmarshal(recommendationsJSON, &recommendations)

	report := &model.TeamWeeklyReport{
		TeamID:             teamID,
		WeekStart:          weekStart.Format("2006-01-02"),
		WeekEnd:            weekStart.AddDate(0, 0, 7).Format("2006-01-02"),
		ExecutiveSummary:   executiveSummary,
		VelocityAnalysis:   velocityAnalysis,
		CollaborationNotes: collaborationNotes,
		TeamBlockers:       make([]*model.TeamBlocker, 0, len(blockers)),
		Recommendations:    recommendations,
		GeneratedAt:        generatedAt.Format(time.RFC3339),
	}

	for _, b := range blockers {
		userIDs := make([]string, len(b.AffectedUsers))
		for i, uid := range b.AffectedUsers {
			userIDs[i] = string(uid)
		}
		report.TeamBlockers = append(report.TeamBlockers, &model.TeamBlocker{
			Title:         b.Title,
			AffectedUsers: userIDs,
			Action:        b.Action,
		})
	}

	return report, nil
}

func (r *queryResolver) fetchTeamMembers(ctx context.Context, teamID string) ([]types.UserID, error) {
	// For now, just get all users. In production, you'd have a team_members table
	query := `SELECT DISTINCT user_id FROM events LIMIT 10`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []types.UserID
	for rows.Next() {
		var userID types.UserID
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		members = append(members, userID)
	}

	return members, nil
}

func (r *queryResolver) fetchUserReport(ctx context.Context, userID string, weekStart time.Time) (*analytics.UserWeeklyReport, error) {
	query := `
        SELECT wins, in_progress, blocked, notes
        FROM weekly_reports
        WHERE user_id = $1 AND week_start = $2
        LIMIT 1
    `

	var winsJSON, progressJSON, blockedJSON []byte
	var notes sql.NullString

	err := r.DB.QueryRowContext(ctx, query, userID, weekStart).Scan(
		&winsJSON, &progressJSON, &blockedJSON, &notes,
	)

	if err != nil {
		return nil, err
	}

	var wins []analytics.WinItem
	var inProgress []analytics.ProgressItem
	var blocked []analytics.BlockedItem
	json.Unmarshal(winsJSON, &wins)
	json.Unmarshal(progressJSON, &inProgress)
	json.Unmarshal(blockedJSON, &blocked)

	report := &analytics.UserWeeklyReport{
		UserID:     types.UserID(userID),
		Wins:       wins,
		InProgress: inProgress,
		Blocked:    blocked,
	}

	if notes.Valid {
		report.Notes = notes.String
	}

	return report, nil
}

func (r *queryResolver) aggregateTeamMetrics(teamID string, week types.TimeRange, userReports map[types.UserID]analytics.UserWeeklyReport) analytics.WeeklyTeamAggregationMetrics {
	metrics := analytics.WeeklyTeamAggregationMetrics{
		TeamID:    types.TeamID(teamID),
		WeekStart: week.Start,
		WeekEnd:   week.End,
	}

	totalWins := 0
	totalBlockers := 0

	for _, report := range userReports {
		totalWins += len(report.Wins)
		totalBlockers += len(report.Blocked)
		metrics.TotalUserCount++
	}

	metrics.TotalWinsReported = analytics.TotalWins(totalWins)
	metrics.TotalBlockersReported = analytics.TotalBlockers(totalBlockers)
	metrics.AverageProductivityScore = analytics.AverageProductivity(75.0)
	metrics.AverageCollaborationScore = analytics.AverageCollaboration(80.0)

	return metrics
}

func (r *queryResolver) buildBasicTeamReport(teamID string, members []types.UserID, week types.TimeRange) *model.TeamWeeklyReport {
	return &model.TeamWeeklyReport{
		TeamID:             teamID,
		WeekStart:          week.Start.Format("2006-01-02"),
		WeekEnd:            week.End.Format("2006-01-02"),
		ExecutiveSummary:   fmt.Sprintf("Team report for %d members", len(members)),
		VelocityAnalysis:   "Analysis pending - insufficient data",
		CollaborationNotes: "Collaboration metrics being collected",
		TeamBlockers:       []*model.TeamBlocker{},
		Recommendations:    []string{"Ensure all team members submit weekly reports"},
		MemberCount:        int32(len(members)),
		GeneratedAt:        time.Now().Format(time.RFC3339),
	}
}

func (r *queryResolver) convertTeamReportToGraphQL(report analytics.TeamWeeklyReport, teamID, weekStart, weekEnd string, memberCount int) *model.TeamWeeklyReport {
	gqlReport := &model.TeamWeeklyReport{
		TeamID:             teamID,
		WeekStart:          weekStart,
		WeekEnd:            weekEnd,
		ExecutiveSummary:   report.ExecutiveSummary,
		VelocityAnalysis:   report.VelocityAnalysis,
		CollaborationNotes: report.CollaborationNotes,
		TeamBlockers:       make([]*model.TeamBlocker, 0, len(report.TeamBlockers)),
		Recommendations:    report.Recommendations,
		MemberCount:        int32(memberCount),
		GeneratedAt:        report.GeneratedAt.Format(time.RFC3339),
	}

	for _, blocker := range report.TeamBlockers {
		userIDs := make([]string, len(blocker.AffectedUsers))
		for i, uid := range blocker.AffectedUsers {
			userIDs[i] = string(uid)
		}
		gqlReport.TeamBlockers = append(gqlReport.TeamBlockers, &model.TeamBlocker{
			Title:         blocker.Title,
			AffectedUsers: userIDs,
			Action:        blocker.Action,
		})
	}

	return gqlReport
}

func (r *queryResolver) saveTeamReport(ctx context.Context, teamID string, week types.TimeRange, report analytics.TeamWeeklyReport) error {
	blockersJSON, _ := json.Marshal(report.TeamBlockers)
	recommendationsJSON, _ := json.Marshal(report.Recommendations)

	query := `
        INSERT INTO team_weekly_reports (
            team_id, week_start, week_end,
            executive_summary, velocity_analysis, collaboration_notes,
            team_blockers, recommendations, generated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (team_id, week_start)
        DO UPDATE SET
            executive_summary = $4,
            velocity_analysis = $5,
            collaboration_notes = $6,
            team_blockers = $7,
            recommendations = $8,
            generated_at = $9
    `

	_, err := r.DB.ExecContext(ctx, query,
		teamID, week.Start, week.End,
		report.ExecutiveSummary, report.VelocityAnalysis, report.CollaborationNotes,
		blockersJSON, recommendationsJSON, time.Now(),
	)
	return err
}
