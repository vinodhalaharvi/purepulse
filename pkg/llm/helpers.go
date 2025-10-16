// ============================================================================
// pkg/llm/helpers.go - Helper Functions
// ============================================================================

package llm

// ============================================================================
// TYPE-SAFE CONSTRUCTORS
// ============================================================================

// NewPrompt creates a new prompt with type-safe fields
func NewPrompt(
	system SystemMessage,
	user UserMessage,
	maxTokens MaxTokens,
	temperature Temperature,
) Prompt {
	return Prompt{}
}

// NewTokenUsage creates token usage record
func NewTokenUsage(input InputTokens, output OutputTokens, cost TokenCost) TokenUsage {
	return TokenUsage{}
}

// NewValidationError creates a validation error
func NewValidationError(field ValidationField, message ValidationMessage, code ValidationCode) ValidationError {
	return ValidationError{}
}

// NewSummaryMetrics creates summary metrics
func NewSummaryMetrics(
	totalEvents EventCount,
	platformsActive PlatformCount,
	productivity ProductivityScore,
	focus FocusScore,
	collaboration CollaborationScore,
) SummaryMetrics {
	return SummaryMetrics{}
}
