package graph

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
	"github.com/vinodhalaharvi/purepulse/pkg/graphql/graph/model"
	"github.com/vinodhalaharvi/purepulse/pkg/llm"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// UserWeekly is the resolver for the userWeekly field.
func (r *queryResolver) UserWeekly(ctx context.Context, userID string, weekStart string, weekEnd string) (*model.UserWeeklyReport, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, fmt.Errorf("invalid weekStart date: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", weekEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid weekEnd date: %w", err)
	}

	// Create time range
	timeRange := types.TimeRange{
		Start: startDate,
		End:   endDate,
	}

	// First, try to fetch existing report from database
	existingReport, err := r.fetchExistingReport(ctx, userID, startDate)
	if err == nil && existingReport != nil {
		log.Printf("Found existing report for user %s", userID)
		return existingReport, nil
	}

	// If no existing report, generate a new one
	log.Printf("Generating new report for user %s from %s to %s", userID, weekStart, weekEnd)

	// Fetch daily activity - using simplified structure
	dailyActivity, err := r.fetchDailyActivity(ctx, userID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch daily activity: %w", err)
	}

	if len(dailyActivity) == 0 {
		// Return empty report if no activity
		notes := "No activity found for this week"
		return &model.UserWeeklyReport{
			UserID:      userID,
			Wins:        []*model.Win{},
			InProgress:  []*model.InProgress{},
			Blocked:     []*model.Blocked{},
			Notes:       &notes,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, nil
	}

	// Initialize Claude client if not already done
	if r.ClaudeClient == nil {
		r.ClaudeClient = &llm.ClaudeClient{
			APIKey:     os.Getenv("ANTHROPIC_API_KEY"),
			Model:      "claude-sonnet-4-20250514",
			BaseURL:    "https://api.anthropic.com/v1",
			MaxRetries: 3,
			Timeout:    60 * time.Second,
		}
	}

	// For now, use basic analysis (you can add Claude later)
	return r.buildBasicReport(userID, dailyActivity), nil
}

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
			PercentComplete: int32(p.PercentComplete), // Convert to int32
			Blocker:         blocker,
			DueDate:         dueDate,
		})
	}

	for _, b := range blocked {
		report.Blocked = append(report.Blocked, &model.Blocked{
			Title:           b.Title,
			BlockedBy:       b.BlockedBy,
			DurationHours:   int32(b.DurationHours), // Convert to int32
			ImpactLevel:     string(b.ImpactLevel),
			SuggestedAction: b.SuggestedAction,
		})
	}

	if notes.Valid {
		report.Notes = &notes.String
	}

	return report, nil
}

// SimpleDailyActivity is a simplified version for now
type SimpleDailyActivity struct {
	UserID     string
	Date       time.Time
	Source     string
	EventCount int
	Channels   pq.StringArray
	TotalSize  int
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

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
