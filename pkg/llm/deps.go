package llm

import (
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
	// TODO: Implement summary merging logic
	return Summary{}
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
