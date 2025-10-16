package llm

import (
	"fmt"

	"github.com/vinodhalaharvi/purekernels/pkg/either"
)

// ============================================================================
// VALIDATION FUNCTIONS (Pure)
// ============================================================================

// validationMonoid is a helper for creating validations
var validationMonoid = ValidationErrorsMonoid{}

// ValidateResponse performs basic response validation
func ValidateResponse(response LLMResponse) LLMValidation {
	errors := []ValidationError{}

	if response.Content == "" {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: "response content is empty",
		})
	}

	if len(errors) > 0 {
		return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, errors)
	}

	return either.Valid[[]ValidationError, LLMResponse](validationMonoid, response)
}

// ValidateContentLength validates content length bounds
func ValidateContentLength(response LLMResponse, minLen, maxLen int) LLMValidation {
	errors := []ValidationError{}

	contentLen := len(response.Content)
	if contentLen < minLen {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: fmt.Sprintf("content too short: %d < %d", contentLen, minLen),
		})
	}

	if contentLen > maxLen {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: fmt.Sprintf("content too long: %d > %d", contentLen, maxLen),
		})
	}

	if len(errors) > 0 {
		return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, errors)
	}

	return either.Valid[[]ValidationError, LLMResponse](validationMonoid, response)
}

// ValidateStopReason validates the stop reason
func ValidateStopReason(response LLMResponse, allowedReasons []string) LLMValidation {
	for _, reason := range allowedReasons {
		if response.StopReason == reason {
			return either.Valid[[]ValidationError, LLMResponse](validationMonoid, response)
		}
	}

	return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, []ValidationError{
		{
			Field:   "stop_reason",
			Message: fmt.Sprintf("invalid stop reason: %s", response.StopReason),
		},
	})
}

// ValidateTokenUsage validates token consumption
func ValidateTokenUsage(response LLMResponse, maxTokens int) LLMValidation {
	if response.Usage.TotalTokens > maxTokens {
		return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, []ValidationError{
			{
				Field:   "token_usage",
				Message: fmt.Sprintf("exceeded max tokens: %d > %d", response.Usage.TotalTokens, maxTokens),
			},
		})
	}

	return either.Valid[[]ValidationError, LLMResponse](validationMonoid, response)
}

// CombineValidations combines multiple validations (parallel)
func CombineValidations(validations ...LLMValidation) LLMValidation {
	// TODO: Implement applicative combination
	if len(validations) == 0 {
		return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, []ValidationError{})
	}
	return validations[0]
}
