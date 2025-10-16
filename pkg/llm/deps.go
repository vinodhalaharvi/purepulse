package llm

import (
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/types"
	"golang.org/x/exp/constraints"
)

// ============================================================================
// PROMPT MONOID
// ============================================================================

// Empty returns the identity prompt
func (pm PromptMonoid) Empty() Prompt {
	return Prompt{
		SystemMessage: "",
		UserMessage:   "",
		Context:       make(map[string]interface{}),
		MaxTokens:     0,
		Temperature:   0.0,
	}
}

// Combine merges two prompts
func (pm PromptMonoid) Combine(a, b Prompt) Prompt {
	return Prompt{
		SystemMessage: a.SystemMessage + "\n\n" + b.SystemMessage,
		UserMessage:   a.UserMessage + "\n\n" + b.UserMessage,
		Context:       MergeMaps(a.Context, b.Context),
		MaxTokens:     Max(a.MaxTokens, b.MaxTokens),
		Temperature:   (a.Temperature + b.Temperature) / 2,
	}
}

// ============================================================================
// VALIDATION ERRORS MONOID
// ============================================================================

// Empty returns empty error list
func (vm ValidationErrorsMonoid) Empty() []ValidationError {
	return []ValidationError{}
}

// Combine appends error lists
func (vm ValidationErrorsMonoid) Combine(a, b []ValidationError) []ValidationError {
	result := make([]ValidationError, 0, len(a)+len(b))
	result = append(result, a...)
	result = append(result, b...)
	return result
}

// ============================================================================
// SUMMARY MONOID
// ============================================================================

// Empty returns the identity summary
func (sm SummaryMonoid) Empty() Summary {
	return Summary{
		Highlights: []string{},
		Insights:   []string{},
		Platforms:  []types.Platform{},
	}
}

// Combine merges two summaries
func (sm SummaryMonoid) Combine(a, b Summary) Summary {
	// Take the first non-empty ID
	id := a.ID
	if id == "" {
		id = b.ID
	}

	// Take the first non-empty UserID
	userID := a.UserID
	if userID == "" {
		userID = b.UserID
	}

	// Take the first non-empty Type
	summaryType := a.Type
	if summaryType == "" {
		summaryType = b.Type
	}

	// Concatenate content with separator
	content := a.Content
	if b.Content != "" {
		if content != "" {
			content += "\n\n---\n\n"
		}
		content += b.Content
	}

	// Merge highlights (deduplicate)
	highlights := mergeUnique(a.Highlights, b.Highlights)

	// Merge insights (deduplicate)
	insights := mergeUnique(a.Insights, b.Insights)

	// Merge platforms (deduplicate)
	platforms := mergePlatforms(a.Platforms, b.Platforms)

	// Combine metrics (sum counters, average scores)
	metrics := SummaryMetrics{
		TotalEvents:        a.Metrics.TotalEvents + b.Metrics.TotalEvents,
		PlatformsActive:    Max(a.Metrics.PlatformsActive, b.Metrics.PlatformsActive),
		ProductivityScore:  (a.Metrics.ProductivityScore + b.Metrics.ProductivityScore) / 2,
		FocusScore:         (a.Metrics.FocusScore + b.Metrics.FocusScore) / 2,
		CollaborationScore: (a.Metrics.CollaborationScore + b.Metrics.CollaborationScore) / 2,
	}

	// Take the more recent timestamp
	generatedAt := a.GeneratedAt
	if b.GeneratedAt.After(a.GeneratedAt) {
		generatedAt = b.GeneratedAt
	}

	// Take the first non-empty model
	modelUsed := a.ModelUsed
	if modelUsed == "" {
		modelUsed = b.ModelUsed
	}

	// Combine token usage
	tokenUsage := TokenUsage{
		InputTokens:  a.TokenUsage.InputTokens + b.TokenUsage.InputTokens,
		OutputTokens: a.TokenUsage.OutputTokens + b.TokenUsage.OutputTokens,
		TotalTokens:  a.TokenUsage.TotalTokens + b.TokenUsage.TotalTokens,
	}

	// Merge time range
	timeRange := types.TimeRange{
		Start: minTime(a.TimeRange.Start, b.TimeRange.Start),
		End:   maxTime(a.TimeRange.End, b.TimeRange.End),
	}

	return Summary{
		ID:          id,
		UserID:      userID,
		Type:        summaryType,
		Content:     content,
		Highlights:  highlights,
		Insights:    insights,
		Metrics:     metrics,
		TimeRange:   timeRange,
		Platforms:   platforms,
		GeneratedAt: generatedAt,
		ModelUsed:   modelUsed,
		TokenUsage:  tokenUsage,
	}
}

// ============================================================================
// LLM RESPONSE LIST MONOID
// ============================================================================

// Empty returns empty response list
func (m LLMResponseListMonoid) Empty() []LLMResponse {
	return []LLMResponse{}
}

// Combine concatenates response lists
func (m LLMResponseListMonoid) Combine(a, b []LLMResponse) []LLMResponse {
	result := make([]LLMResponse, 0, len(a)+len(b))
	result = append(result, a...)
	result = append(result, b...)
	return result
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// MergeMaps combines two maps (right-biased)
func MergeMaps[K comparable, V any](a, b map[K]V) map[K]V {
	result := make(map[K]V, len(a)+len(b))
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}

// Max returns the maximum of two ordered values
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum of two ordered values
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// mergeUnique combines two string slices and removes duplicates
func mergeUnique(a, b []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range a {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	for _, item := range b {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// mergePlatforms combines two platform slices and removes duplicates
func mergePlatforms(a, b []types.Platform) []types.Platform {
	seen := make(map[types.Platform]bool)
	result := []types.Platform{}

	for _, platform := range a {
		if !seen[platform] {
			seen[platform] = true
			result = append(result, platform)
		}
	}

	for _, platform := range b {
		if !seen[platform] {
			seen[platform] = true
			result = append(result, platform)
		}
	}

	return result
}

// minTime returns the earlier of two times
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// maxTime returns the later of two times
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
