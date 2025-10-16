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

	if response.StopReason == "" {
		errors = append(errors, ValidationError{
			Field:   "stop_reason",
			Message: "stop reason is missing",
		})
	}

	if response.ModelVersion == "" {
		errors = append(errors, ValidationError{
			Field:   "model_version",
			Message: "model version is missing",
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
	// If no allowed reasons specified, accept any non-empty reason
	if len(allowedReasons) == 0 {
		if response.StopReason != "" {
			return either.Valid[[]ValidationError, LLMResponse](validationMonoid, response)
		}
		return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, []ValidationError{
			{
				Field:   "stop_reason",
				Message: "stop reason is empty",
			},
		})
	}

	// Check against allowed reasons
	for _, reason := range allowedReasons {
		if response.StopReason == reason {
			return either.Valid[[]ValidationError, LLMResponse](validationMonoid, response)
		}
	}

	return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, []ValidationError{
		{
			Field:   "stop_reason",
			Message: fmt.Sprintf("invalid stop reason: %s (allowed: %v)", response.StopReason, allowedReasons),
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

	if response.Usage.TotalTokens < 1 {
		return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, []ValidationError{
			{
				Field:   "token_usage",
				Message: "token usage is zero or negative",
			},
		})
	}

	return either.Valid[[]ValidationError, LLMResponse](validationMonoid, response)
}

// ValidateJSONContent validates that content is valid JSON
func ValidateJSONContent(response LLMResponse) LLMValidation {
	// Try to parse as SummaryJSON
	parseResult := ParseJSON[SummaryJSON](response.Content)

	if !parseResult.IsOk() {
		return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, []ValidationError{
			{
				Field:   "content",
				Message: fmt.Sprintf("content is not valid JSON: %v", parseResult.Error()),
			},
		})
	}

	return either.Valid[[]ValidationError, LLMResponse](validationMonoid, response)
}

// CombineValidations combines multiple validations (parallel - accumulates all errors)
func CombineValidations(validations ...LLMValidation) LLMValidation {
	if len(validations) == 0 {
		// Empty list is invalid
		return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, []ValidationError{
			{
				Field:   "validations",
				Message: "no validations provided",
			},
		})
	}

	// Accumulate all errors (applicative composition)
	allErrors := []ValidationError{}
	var lastValidResponse LLMResponse

	for _, validation := range validations {
		if validation.IsValid() {
			// Keep track of valid response
			lastValidResponse = validation.GetValue()
		} else {
			// Accumulate errors
			allErrors = append(allErrors, validation.GetErrors()...)
		}
	}

	// If any validation failed, return all errors
	if len(allErrors) > 0 {
		return either.Invalid[[]ValidationError, LLMResponse](validationMonoid, allErrors)
	}

	// All validations passed
	return either.Valid[[]ValidationError, LLMResponse](validationMonoid, lastValidResponse)
}

// ============================================================================
// COMPOSITE VALIDATORS (High-Level)
// ============================================================================

// ValidateFullResponse performs all standard validations
func ValidateFullResponse(response LLMResponse) LLMValidation {
	return CombineValidations(
		ValidateResponse(response),
		ValidateContentLength(response, 10, 10000), // Min 10 chars, max 10k
		ValidateStopReason(response, []string{"end_turn", "max_tokens", "stop_sequence"}),
		ValidateTokenUsage(response, 5000), // Max 5k tokens
		ValidateJSONContent(response),
	)
}

// ValidateBasicResponse performs minimal validations (for testing/development)
func ValidateBasicResponse(response LLMResponse) LLMValidation {
	return CombineValidations(
		ValidateResponse(response),
		ValidateJSONContent(response),
	)
}

// ValidateWithCustomLimits validates with custom token and length limits
func ValidateWithCustomLimits(response LLMResponse, minLen, maxLen, maxTokens int) LLMValidation {
	return CombineValidations(
		ValidateResponse(response),
		ValidateContentLength(response, minLen, maxLen),
		ValidateTokenUsage(response, maxTokens),
		ValidateJSONContent(response),
	)
}
