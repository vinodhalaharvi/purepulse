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

const jsonInstructions = `
You must respond with valid JSON in the following format:
{
  "summary": "Brief overview of the user's activity",
  "highlights": ["Key achievement 1", "Key achievement 2", "..."],
  "insights": ["Insight about patterns 1", "Insight 2", "..."],
  "metrics": {
    "total_events": 0,
    "platforms_active": 0,
    "productivity_score": 0.0,
    "focus_score": 0.0,
    "collaboration_score": 0.0
  }
}

Scores should be between 0.0 and 1.0. Do not include any text outside the JSON.`

// DailySummaryPrompt creates a daily summary prompt
func DailySummaryPrompt(
	userID types.UserID,
	eventsList []events.Event,
	date time.Time,
) Prompt {
	return Prompt{
		SystemMessage: "You are an AI assistant that creates daily activity summaries. " + jsonInstructions,
		UserMessage: fmt.Sprintf(
			"Summarize activities for user '%s' on %s. They had %d events.",
			userID,
			date.Format("2006-01-02"),
			len(eventsList),
		),
		Context: map[string]interface{}{
			"user_id":     userID,
			"date":        date,
			"event_count": len(eventsList),
			"events":      formatEventsForPrompt(eventsList),
		},
		MaxTokens:   1500,
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
		SystemMessage: "You are an AI assistant that creates weekly activity summaries. " + jsonInstructions,
		UserMessage: fmt.Sprintf(
			"Summarize activities for user '%s' from %s to %s. They had %d events total.",
			userID,
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
			len(eventsList),
		),
		Context: map[string]interface{}{
			"user_id":     userID,
			"start_date":  startDate,
			"end_date":    endDate,
			"event_count": len(eventsList),
			"events":      formatEventsForPrompt(eventsList),
		},
		MaxTokens:   2500,
		Temperature: 0.7,
	}
}

// ActivityInsightsPrompt creates an insights prompt
func ActivityInsightsPrompt(
	userID types.UserID,
	activity events.UserActivity,
) Prompt {
	return Prompt{
		SystemMessage: "You are an AI assistant that provides productivity insights. " + jsonInstructions,
		UserMessage: fmt.Sprintf(
			"Provide insights on activity patterns for user '%s'. Total events: %d across platforms.",
			userID,
			activity.TotalEvents(),
		),
		Context: map[string]interface{}{
			"user_id":      userID,
			"total_events": activity.TotalEvents(),
			"slack_count":  len(activity.Slack),
			"github_count": len(activity.GitHub),
			"jira_count":   len(activity.Jira),
			"zoom_count":   len(activity.Zoom),
		},
		MaxTokens:   2000,
		Temperature: 0.7,
	}
}

// ProductivityAnalysisPrompt creates a productivity analysis prompt
func ProductivityAnalysisPrompt(
	userID types.UserID,
	activity events.UserActivity,
) Prompt {
	return Prompt{
		SystemMessage: "You are an AI assistant that analyzes productivity metrics. " + jsonInstructions,
		UserMessage: fmt.Sprintf(
			"Analyze productivity metrics for user '%s'. Focus on patterns, trends, and recommendations.",
			userID,
		),
		Context: map[string]interface{}{
			"user_id":      userID,
			"total_events": activity.TotalEvents(),
			"slack_count":  len(activity.Slack),
			"github_count": len(activity.GitHub),
			"jira_count":   len(activity.Jira),
			"zoom_count":   len(activity.Zoom),
		},
		MaxTokens:   2000,
		Temperature: 0.7,
	}
}

// CrossPlatformCorrelationPrompt creates a correlation prompt
func CrossPlatformCorrelationPrompt(
	userID types.UserID,
	eventsList []events.Event,
) Prompt {
	return Prompt{
		SystemMessage: "You are an AI assistant that finds cross-platform patterns. " + jsonInstructions,
		UserMessage: fmt.Sprintf(
			"Find correlations in activities for user '%s'. Analyze %d events across different platforms.",
			userID,
			len(eventsList),
		),
		Context: map[string]interface{}{
			"user_id":     userID,
			"event_count": len(eventsList),
			"events":      formatEventsForPrompt(eventsList),
		},
		MaxTokens:   2500,
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

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// formatEventsForPrompt creates a concise event summary for the prompt
func formatEventsForPrompt(eventsList []events.Event) string {
	if len(eventsList) == 0 {
		return "No events"
	}

	// Group by platform and type
	summary := make(map[string]map[string]int)

	for _, event := range eventsList {
		platform := string(event.Source)
		eventType := string(event.Type)

		if summary[platform] == nil {
			summary[platform] = make(map[string]int)
		}
		summary[platform][eventType]++
	}

	// Format as string (concise for prompt)
	result := ""
	for platform, types := range summary {
		result += fmt.Sprintf("\n%s: ", platform)
		for eventType, count := range types {
			result += fmt.Sprintf("%d %s, ", count, eventType)
		}
	}

	return result
}
