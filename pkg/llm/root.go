package llm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/effect"
	"github.com/vinodhalaharvi/purekernels/pkg/result"

	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// PROMPT BUILDER METHODS
// ============================================================================

// Contramap transforms the input before building
func (pb PromptBuilder[A]) Contramap(f func(A) A) PromptBuilder[A] {
	return PromptBuilder[A]{
		Build: func(a A) Prompt {
			return pb.Build(f(a))
		},
	}
}

// Compose chains two prompt builders
func (pb PromptBuilder[A]) Compose(other PromptBuilder[Prompt]) PromptBuilder[A] {
	return PromptBuilder[A]{
		Build: func(a A) Prompt {
			intermediate := pb.Build(a)
			return other.Build(intermediate)
		},
	}
}

// ============================================================================
// LLM EFFECT METHODS
// ============================================================================

// ============================================================================
// PIPELINE METHODS
// ============================================================================

// NewPipeline creates a new pipeline
func NewPipeline[A, B any](f func(A) result.Result[B]) Pipeline[A, B] {
	return Pipeline[A, B]{
		Run: f,
	}
}

// ============================================================================
// PIPELINE COMPOSITION (Standalone Functions)
// ============================================================================

// PipelineThen adds a step to the pipeline
func PipelineThen[A, B, C any](
	p Pipeline[A, B],
	f func(B) result.Result[C],
) Pipeline[A, C] {
	return Pipeline[A, C]{
		Run: func(a A) result.Result[C] {
			resB := p.Run(a)
			if !resB.IsOk() {
				return result.Err[C](resB.Error())
			}
			return f(resB.Unwrap())
		},
	}
}

// ComposePipelines composes two pipelines
func ComposePipelines[A, B, C any](
	p1 Pipeline[A, B],
	p2 Pipeline[B, C],
) Pipeline[A, C] {
	return Pipeline[A, C]{
		Run: func(a A) result.Result[C] {
			resB := p1.Run(a)
			if !resB.IsOk() {
				return result.Err[C](resB.Error())
			}
			return p2.Run(resB.Unwrap())
		},
	}
}

// ============================================================================
// SUMMARY REPOSITORY METHODS
// ============================================================================

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func formatStringSlice(slice []string) string {
	// TODO: Format as JSON or PostgreSQL array
	return fmt.Sprintf("%v", slice)
}

func formatPlatformSlice(slice []types.Platform) string {
	// TODO: Format as JSON or PostgreSQL array
	return fmt.Sprintf("%v", slice)
}

// ============================================================================
// LLM EFFECT METHODS
// ============================================================================

// Map transforms the success value
func (e LLMEffect[A]) Map(f func(A) A) LLMEffect[A] {
	return LLMEffect[A]{
		Run: func(ctx context.Context) effect.Writer[[]string, result.Result[A]] {
			writer := e.Run(ctx)
			resA, logs := writer.Run()

			if !resA.IsOk() {
				return effect.NewWriter(resA, logs) // ← Value first, log second
			}

			transformed := f(resA.Unwrap())
			return effect.NewWriter(result.Ok(transformed), logs) // ← Value first, log second
		},
	}
}

// Then sequences effects
func (e LLMEffect[A]) Then(next LLMEffect[A]) LLMEffect[A] {
	return LLMEffect[A]{
		Run: func(ctx context.Context) effect.Writer[[]string, result.Result[A]] {
			// Run first effect
			writer1 := e.Run(ctx)
			res1, logs1 := writer1.Run()

			if !res1.IsOk() {
				return effect.NewWriter(res1, logs1) // ← Value first, log second
			}

			// Run second effect
			writer2 := next.Run(ctx)
			res2, logs2 := writer2.Run()

			// Combine logs
			combinedLogs := append(logs1, logs2...)
			return effect.NewWriter(res2, combinedLogs) // ← Value first, log second
		},
	}
}

// MapLogs transforms the log accumulator
func (e LLMEffect[A]) MapLogs(f func([]string) []string) LLMEffect[A] {
	return LLMEffect[A]{
		Run: func(ctx context.Context) effect.Writer[[]string, result.Result[A]] {
			writer := e.Run(ctx)
			resA, logs := writer.Run()
			transformedLogs := f(logs)
			return effect.NewWriter(resA, transformedLogs) // ← Value first, log second
		},
	}
}

// ============================================================================
// SUMMARY GENERATOR METHODS
// ============================================================================

// GenerateDailySummary: full pipeline for daily summary
func (sg *SummaryGenerator) GenerateDailySummary(
	ctx context.Context,
	userID types.UserID,
	eventsList []events.Event,
	date time.Time,
) effect.Writer[[]string, result.Result[Summary]] {
	logs := []string{fmt.Sprintf("generating_daily_summary: user=%s, date=%s", userID, date.Format("2006-01-02"))}

	// Build prompt
	prompt := DailySummaryPrompt(userID, eventsList, date)
	logs = append(logs, "prompt_built")

	// Generate with LLM
	response, err := sg.Client.Generate(ctx, prompt)
	if err != nil {
		logs = append(logs, fmt.Sprintf("llm_failed: %v", err))
		return effect.NewWriter(result.Err[Summary](err), logs) // ← Value first, log second
	}
	logs = append(logs, "llm_succeeded")

	// Validate response
	validation := sg.Validator(response)
	if !validation.IsValid() {
		logs = append(logs, fmt.Sprintf("validation_failed: %v", validation.GetErrors()))
		return effect.NewWriter(
			result.Err[Summary](fmt.Errorf("validation failed: %v", validation.GetErrors())),
			logs, // ← Value first, log second
		)
	}
	logs = append(logs, "validation_passed")

	// Parse response
	summaryResult := sg.ResponseParser(response)
	if !summaryResult.IsOk() {
		logs = append(logs, fmt.Sprintf("parsing_failed: %v", summaryResult.Error()))
		return effect.NewWriter(summaryResult, logs) // ← Value first, log second
	}
	logs = append(logs, "parsing_succeeded")

	summary := summaryResult.Unwrap()
	summary.UserID = userID
	summary.TimeRange = types.TimeRange{Start: date, End: date.Add(24 * time.Hour)}

	return effect.NewWriter(result.Ok(summary), logs) // ← Value first, log second
}

// GenerateWeeklySummary: full pipeline for weekly summary
func (sg *SummaryGenerator) GenerateWeeklySummary(
	ctx context.Context,
	userID types.UserID,
	eventsList []events.Event,
	startDate, endDate time.Time,
) effect.Writer[[]string, result.Result[Summary]] {
	logs := []string{fmt.Sprintf("generating_weekly_summary: user=%s, range=%s to %s",
		userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))}

	// Build prompt
	prompt := WeeklySummaryPrompt(userID, eventsList, startDate, endDate)
	logs = append(logs, "prompt_built")

	// Generate with LLM
	response, err := sg.Client.Generate(ctx, prompt)
	if err != nil {
		logs = append(logs, fmt.Sprintf("llm_failed: %v", err))
		return effect.NewWriter(result.Err[Summary](err), logs) // ← Value first, log second
	}
	logs = append(logs, "llm_succeeded")

	// Validate response
	validation := sg.Validator(response)
	if !validation.IsValid() {
		logs = append(logs, fmt.Sprintf("validation_failed: %v", validation.GetErrors()))
		return effect.NewWriter(
			result.Err[Summary](fmt.Errorf("validation failed: %v", validation.GetErrors())),
			logs, // ← Value first, log second
		)
	}
	logs = append(logs, "validation_passed")

	// Parse response
	summaryResult := sg.ResponseParser(response)
	if !summaryResult.IsOk() {
		logs = append(logs, fmt.Sprintf("parsing_failed: %v", summaryResult.Error()))
		return effect.NewWriter(summaryResult, logs) // ← Value first, log second
	}
	logs = append(logs, "parsing_succeeded")

	summary := summaryResult.Unwrap()
	summary.UserID = userID
	summary.TimeRange = types.TimeRange{Start: startDate, End: endDate}

	return effect.NewWriter(result.Ok(summary), logs) // ← Value first, log second
}

// GenerateProductivityInsights: full pipeline for productivity analysis
func (sg *SummaryGenerator) GenerateProductivityInsights(
	ctx context.Context,
	userID types.UserID,
	activity events.UserActivity,
) effect.Writer[[]string, result.Result[Summary]] {
	logs := []string{fmt.Sprintf("generating_productivity_insights: user=%s", userID)}

	// Build prompt
	prompt := ProductivityAnalysisPrompt(userID, activity)
	logs = append(logs, "prompt_built")

	// Generate with LLM
	response, err := sg.Client.Generate(ctx, prompt)
	if err != nil {
		logs = append(logs, fmt.Sprintf("llm_failed: %v", err))
		return effect.NewWriter(result.Err[Summary](err), logs) // ← Value first, log second
	}
	logs = append(logs, "llm_succeeded")

	// Validate response
	validation := sg.Validator(response)
	if !validation.IsValid() {
		logs = append(logs, fmt.Sprintf("validation_failed: %v", validation.GetErrors()))
		return effect.NewWriter(
			result.Err[Summary](fmt.Errorf("validation failed: %v", validation.GetErrors())),
			logs, // ← Value first, log second
		)
	}
	logs = append(logs, "validation_passed")

	// Parse response
	summaryResult := sg.ResponseParser(response)
	if !summaryResult.IsOk() {
		logs = append(logs, fmt.Sprintf("parsing_failed: %v", summaryResult.Error()))
		return effect.NewWriter(summaryResult, logs) // ← Value first, log second
	}
	logs = append(logs, "parsing_succeeded")

	summary := summaryResult.Unwrap()
	summary.UserID = userID
	summary.AISummary.Type = SummaryTypeProductivity

	return effect.NewWriter(result.Ok(summary), logs) // ← Value first, log second
}

// GenerateAllSummaries: parallel generation of all summary types
func (sg *SummaryGenerator) GenerateAllSummaries(
	ctx context.Context,
	userID types.UserID,
	eventsList []events.Event,
	timeRange types.TimeRange,
) effect.Writer[[]string, result.Result[[]Summary]] {
	logs := []string{fmt.Sprintf("generating_all_summaries: user=%s", userID)}

	// Generate daily summary
	dailyWriter := sg.GenerateDailySummary(ctx, userID, eventsList, timeRange.Start)
	dailyResult, dailyLogs := dailyWriter.Run()
	logs = append(logs, dailyLogs...)

	// Generate weekly summary
	weeklyWriter := sg.GenerateWeeklySummary(ctx, userID, eventsList, timeRange.Start, timeRange.End)
	weeklyResult, weeklyLogs := weeklyWriter.Run()
	logs = append(logs, weeklyLogs...)

	// Collect results
	summaries := []Summary{}

	if dailyResult.IsOk() {
		summaries = append(summaries, dailyResult.Unwrap())
	} else {
		logs = append(logs, fmt.Sprintf("daily_summary_failed: %v", dailyResult.Error()))
	}

	if weeklyResult.IsOk() {
		summaries = append(summaries, weeklyResult.Unwrap())
	} else {
		logs = append(logs, fmt.Sprintf("weekly_summary_failed: %v", weeklyResult.Error()))
	}

	if len(summaries) == 0 {
		return effect.NewWriter(
			result.Err[[]Summary](fmt.Errorf("all summary generations failed")),
			logs, // ← Value first, log second
		)
	}

	return effect.NewWriter(result.Ok(summaries), logs) // ← Value first, log second
}

// ============================================================================
// SUMMARY REPOSITORY METHODS
// ============================================================================

// Save persists a summary to the database
func (sr *SummaryRepository) Save(ctx context.Context, summary Summary) error {
	query := `
		INSERT INTO summaries (
			user_id,
			time_range_start,
			time_range_end,
			activity,
			metrics,
			correlations,
			ai_summary,
			ai_model,
			ai_tokens,
			ai_latency_ms,
			audit_log,
			version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	// Convert to JSONB-compatible structures
	activityJSON := map[string]interface{}{
		"total_events": summary.Activity.TotalEvents,
		"platforms":    summary.Activity.Platforms,
	}

	metricsJSON := map[string]interface{}{
		"total_events":        summary.Metrics.TotalEvents,
		"platforms_active":    summary.Metrics.PlatformsActive,
		"productivity_score":  summary.Metrics.ProductivityScore,
		"focus_score":         summary.Metrics.FocusScore,
		"collaboration_score": summary.Metrics.CollaborationScore,
	}

	correlationsJSON := []interface{}{}

	aiSummaryJSON := map[string]interface{}{
		"type":       summary.AISummary.Type,
		"content":    summary.AISummary.Content,
		"highlights": summary.AISummary.Highlights,
		"insights":   summary.AISummary.Insights,
	}

	auditLog := summary.AuditLog

	var id string
	err := sr.DB.QueryRowContext(ctx, query,
		summary.UserID,
		summary.TimeRangeStart,
		summary.TimeRangeEnd,
		activityJSON,
		metricsJSON,
		correlationsJSON,
		aiSummaryJSON,
		summary.AIModel,
		summary.AITokens,
		summary.AILatencyMS,
		auditLog,
		summary.Version,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to save summary: %w", err)
	}

	// Update summary with generated ID
	summary.ID = id

	return nil
}

// SaveBatch persists multiple summaries
func (sr *SummaryRepository) SaveBatch(ctx context.Context, summaries []Summary) error {
	tx, err := sr.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, summary := range summaries {
		if err := sr.Save(ctx, summary); err != nil {
			return fmt.Errorf("failed to save summary in batch: %w", err)
		}
	}

	return tx.Commit()
}

// FindByUser retrieves summaries for a user within a time range
func (sr *SummaryRepository) FindByUser(
	ctx context.Context,
	userID types.UserID,
	summaryType SummaryType,
	timeRange types.TimeRange,
) ([]Summary, error) {
	query := `
		SELECT 
			id,
			user_id,
			time_range_start,
			time_range_end,
			ai_summary->'type' as type,
			ai_summary->'content' as content,
			ai_summary->'highlights' as highlights,
			ai_summary->'insights' as insights,
			metrics,
			ai_model,
			ai_tokens,
			generated_at
		FROM summaries
		WHERE user_id = $1 
			AND ai_summary->>'type' = $2
			AND time_range_start >= $3 
			AND time_range_end <= $4
		ORDER BY generated_at DESC
	`

	rows, err := sr.DB.QueryContext(ctx, query,
		userID,
		summaryType,
		timeRange.Start,
		timeRange.End,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query summaries: %w", err)
	}
	defer rows.Close()

	summaries := []Summary{}
	for rows.Next() {
		var s Summary
		var typeJSON, contentJSON, highlightsJSON, insightsJSON, metricsJSON []byte

		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.TimeRangeStart,
			&s.TimeRangeEnd,
			&typeJSON,
			&contentJSON,
			&highlightsJSON,
			&insightsJSON,
			&metricsJSON,
			&s.AIModel,
			&s.AITokens,
			&s.GeneratedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan summary: %w", err)
		}

		// Parse JSON fields into AISummary
		var summaryTypeStr string
		json.Unmarshal(typeJSON, &summaryTypeStr)
		s.AISummary.Type = SummaryType(summaryTypeStr)

		json.Unmarshal(contentJSON, &s.AISummary.Content)
		json.Unmarshal(highlightsJSON, &s.AISummary.Highlights)
		json.Unmarshal(insightsJSON, &s.AISummary.Insights)

		// Parse metrics
		var metricsMap map[string]interface{}
		if err := json.Unmarshal(metricsJSON, &metricsMap); err == nil {
			if totalEvents, ok := metricsMap["total_events"].(float64); ok {
				s.Metrics.TotalEvents = int(totalEvents)
			}
			if platformsActive, ok := metricsMap["platforms_active"].(float64); ok {
				s.Metrics.PlatformsActive = int(platformsActive)
			}
			if prodScore, ok := metricsMap["productivity_score"].(float64); ok {
				s.Metrics.ProductivityScore = prodScore
			}
			if focusScore, ok := metricsMap["focus_score"].(float64); ok {
				s.Metrics.FocusScore = focusScore
			}
			if collabScore, ok := metricsMap["collaboration_score"].(float64); ok {
				s.Metrics.CollaborationScore = collabScore
			}
		}

		// Set convenience field
		s.TimeRange = types.TimeRange{
			Start: s.TimeRangeStart,
			End:   s.TimeRangeEnd,
		}

		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating summaries: %w", err)
	}

	return summaries, nil
}

// FindLatest gets the most recent summary for a user and type
func (sr *SummaryRepository) FindLatest(
	ctx context.Context,
	userID types.UserID,
	summaryType SummaryType,
) (*Summary, error) {
	query := `
		SELECT 
			id,
			user_id,
			time_range_start,
			time_range_end,
			ai_summary->'type' as type,
			ai_summary->'content' as content,
			ai_summary->'highlights' as highlights,
			ai_summary->'insights' as insights,
			metrics,
			ai_model,
			ai_tokens,
			generated_at
		FROM summaries
		WHERE user_id = $1 
			AND ai_summary->>'type' = $2
		ORDER BY generated_at DESC
		LIMIT 1
	`

	var s Summary
	var typeJSON, contentJSON, highlightsJSON, insightsJSON, metricsJSON []byte

	err := sr.DB.QueryRowContext(ctx, query, userID, summaryType).Scan(
		&s.ID,
		&s.UserID,
		&s.TimeRangeStart,
		&s.TimeRangeEnd,
		&typeJSON,
		&contentJSON,
		&highlightsJSON,
		&insightsJSON,
		&metricsJSON,
		&s.AIModel,
		&s.AITokens,
		&s.GeneratedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest summary: %w", err)
	}

	// Parse JSON fields into AISummary
	var summaryTypeStr string
	json.Unmarshal(typeJSON, &summaryTypeStr)
	s.AISummary.Type = SummaryType(summaryTypeStr)

	json.Unmarshal(contentJSON, &s.AISummary.Content)
	json.Unmarshal(highlightsJSON, &s.AISummary.Highlights)
	json.Unmarshal(insightsJSON, &s.AISummary.Insights)

	// Parse metrics
	var metricsMap map[string]interface{}
	if err := json.Unmarshal(metricsJSON, &metricsMap); err == nil {
		if totalEvents, ok := metricsMap["total_events"].(float64); ok {
			s.Metrics.TotalEvents = int(totalEvents)
		}
		if platformsActive, ok := metricsMap["platforms_active"].(float64); ok {
			s.Metrics.PlatformsActive = int(platformsActive)
		}
		if prodScore, ok := metricsMap["productivity_score"].(float64); ok {
			s.Metrics.ProductivityScore = prodScore
		}
		if focusScore, ok := metricsMap["focus_score"].(float64); ok {
			s.Metrics.FocusScore = focusScore
		}
		if collabScore, ok := metricsMap["collaboration_score"].(float64); ok {
			s.Metrics.CollaborationScore = collabScore
		}
	}

	// Set convenience field
	s.TimeRange = types.TimeRange{
		Start: s.TimeRangeStart,
		End:   s.TimeRangeEnd,
	}

	return &s, nil
}
