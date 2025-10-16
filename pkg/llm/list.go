package llm

import (
	"fmt"
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// PROMPT TEMPLATES (Pure Functions)
// ============================================================================

// DailySummaryPrompt creates a daily summary prompt
func DailySummaryPrompt(
	userID types.UserID,
	eventsList []events.Event,
	date time.Time,
) Prompt {
	return Prompt{
		SystemMessage: "You are an AI assistant that creates daily activity summaries.",
		UserMessage:   fmt.Sprintf("Summarize activities for %s on %s", userID, date.Format("2006-01-02")),
		Context: map[string]interface{}{
			"user_id":     userID,
			"date":        date,
			"event_count": len(eventsList),
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}
}

// WeeklySummaryPrompt creates a weekly summary prompt
func WeeklySummaryPrompt(
	userID types.UserID,
	eventsList []events.Event,
	startDate, endDate time.Time,
) Prompt {
	return Prompt{
		SystemMessage: "You are an AI assistant that creates weekly activity summaries.",
		UserMessage: fmt.Sprintf(
			"Summarize activities for %s from %s to %s",
			userID,
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
		),
		Context: map[string]interface{}{
			"user_id":     userID,
			"start_date":  startDate,
			"end_date":    endDate,
			"event_count": len(eventsList),
		},
		MaxTokens:   2000,
		Temperature: 0.7,
	}
}

// ActivityInsightsPrompt creates an insights prompt
func ActivityInsightsPrompt(
	userID types.UserID,
	activity events.UserActivity,
) Prompt {
	return Prompt{
		SystemMessage: "You are an AI assistant that provides productivity insights.",
		UserMessage:   fmt.Sprintf("Provide insights on activity patterns for %s", userID),
		Context: map[string]interface{}{
			"user_id":      userID,
			"total_events": activity.TotalEvents(),
		},
		MaxTokens:   1500,
		Temperature: 0.7,
	}
}

// ProductivityAnalysisPrompt creates a productivity analysis prompt
// ProductivityAnalysisPrompt creates a productivity analysis prompt
func ProductivityAnalysisPrompt(
	userID types.UserID,
	activity events.UserActivity, // Changed from events.Metrics
) Prompt {
	return Prompt{
		SystemMessage: "You are an AI assistant that analyzes productivity metrics.",
		UserMessage:   fmt.Sprintf("Analyze productivity metrics for %s", userID),
		Context: map[string]interface{}{
			"user_id":      userID,
			"total_events": activity.TotalEvents(),
		},
		MaxTokens:   1500,
		Temperature: 0.7,
	}
}

// CrossPlatformCorrelationPrompt creates a correlation prompt
func CrossPlatformCorrelationPrompt(
	userID types.UserID,
	eventsList []events.Event,
) Prompt {
	return Prompt{
		SystemMessage: "You are an AI assistant that finds cross-platform patterns.",
		UserMessage:   fmt.Sprintf("Find correlations in activities for %s", userID),
		Context: map[string]interface{}{
			"user_id":     userID,
			"event_count": len(eventsList),
		},
		MaxTokens:   2000,
		Temperature: 0.7,
	}
}

// CombinePrompts combines multiple prompts using PromptMonoid
func CombinePrompts(prompts ...Prompt) Prompt {
	monoid := PromptMonoid{}
	result := monoid.Empty()
	for _, p := range prompts {
		result = monoid.Combine(result, p)
	}
	return result
}
