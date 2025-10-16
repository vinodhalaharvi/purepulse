package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/result"
)

// ============================================================================
// RESPONSE PARSING (Pure Functions - JSON Only)
// ============================================================================

// SummaryJSON is the expected JSON structure from Claude
type SummaryJSON struct {
	Summary    string      `json:"summary"`
	Highlights []string    `json:"highlights"`
	Insights   []string    `json:"insights"`
	Metrics    MetricsJSON `json:"metrics"`
}

// MetricsJSON is the metrics portion of the response
type MetricsJSON struct {
	TotalEvents        int     `json:"total_events"`
	PlatformsActive    int     `json:"platforms_active"`
	ProductivityScore  float64 `json:"productivity_score"`
	FocusScore         float64 `json:"focus_score"`
	CollaborationScore float64 `json:"collaboration_score"`
}

// ParseDailySummary extracts structured summary from LLM response
func ParseDailySummary(response LLMResponse) result.Result[Summary] {
	return parseSummaryJSON(response, SummaryTypeDaily)
}

// ParseWeeklySummary extracts weekly summary
func ParseWeeklySummary(response LLMResponse) result.Result[Summary] {
	return parseSummaryJSON(response, SummaryTypeWeekly)
}

// ParseInsights extracts insights from response
func ParseInsights(response LLMResponse) result.Result[[]string] {
	parsed := ParseJSON[SummaryJSON](response.Content)
	if !parsed.IsOk() {
		return result.Err[[]string](parsed.Error())
	}

	return result.Ok(parsed.Unwrap().Insights)
}

// ParseMetrics extracts metrics from response
func ParseMetrics(response LLMResponse) result.Result[SummaryMetrics] {
	parsed := ParseJSON[SummaryJSON](response.Content)
	if !parsed.IsOk() {
		return result.Err[SummaryMetrics](parsed.Error())
	}

	metricsJSON := parsed.Unwrap().Metrics
	return result.Ok(SummaryMetrics{
		TotalEvents:        metricsJSON.TotalEvents,
		PlatformsActive:    metricsJSON.PlatformsActive,
		ProductivityScore:  metricsJSON.ProductivityScore,
		FocusScore:         metricsJSON.FocusScore,
		CollaborationScore: metricsJSON.CollaborationScore,
	})
}

// ParseJSON extracts typed JSON from content
func ParseJSON[T any](content string) result.Result[T] {
	// Clean content - sometimes LLM adds markdown code blocks
	content = cleanJSONContent(content)

	var value T
	err := json.Unmarshal([]byte(content), &value)
	if err != nil {
		return result.Err[T](fmt.Errorf("failed to parse JSON: %w", err))
	}
	return result.Ok(value)
}

// ============================================================================
// HELPER FUNCTIONS (Pure)
// ============================================================================

// parseSummaryJSON is the common parser for all summary types
func parseSummaryJSON(response LLMResponse, summaryType SummaryType) result.Result[Summary] {
	parsed := ParseJSON[SummaryJSON](response.Content)
	if !parsed.IsOk() {
		return result.Err[Summary](parsed.Error())
	}

	summaryJSON := parsed.Unwrap()

	summary := Summary{
		ID:         generateSummaryID(),
		Type:       summaryType,
		Content:    summaryJSON.Summary,
		Highlights: summaryJSON.Highlights,
		Insights:   summaryJSON.Insights,
		Metrics: SummaryMetrics{
			TotalEvents:        summaryJSON.Metrics.TotalEvents,
			PlatformsActive:    summaryJSON.Metrics.PlatformsActive,
			ProductivityScore:  summaryJSON.Metrics.ProductivityScore,
			FocusScore:         summaryJSON.Metrics.FocusScore,
			CollaborationScore: summaryJSON.Metrics.CollaborationScore,
		},
		GeneratedAt: response.Timestamp,
		ModelUsed:   response.ModelVersion,
		TokenUsage:  response.Usage,
	}

	return result.Ok(summary)
}

// cleanJSONContent removes markdown code blocks and whitespace
func cleanJSONContent(content string) string {
	// Remove ```json and ``` markers
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}

// generateSummaryID creates a unique ID for a summary
func generateSummaryID() string {
	return fmt.Sprintf("summary-%d", time.Now().UnixNano())
}
