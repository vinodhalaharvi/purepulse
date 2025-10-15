package collectors

import (
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/either"
	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purepulse/internal/validation"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// VALIDATION (Validation Applicative)
// ============================================================================

// ValidateFetchRequest validates fetch parameters before executing
func ValidateFetchRequest(
	userID types.UserID,
	timeRange types.TimeRange,
) either.Validation[[]validation.ValidationError, FetchRequest] {

	errorMonoid := monoid.NewListMonoid[validation.ValidationError]()

	// Validate userID
	userIDValidation := validateUserID(userID)

	// Validate time range
	timeRangeValidation := validateTimeRange(timeRange)

	// Combine validations using Validation applicative
	return either.Ap2(
		errorMonoid,
		func(uid types.UserID, tr types.TimeRange) FetchRequest {
			return FetchRequest{
				UserID:    uid,
				TimeRange: tr,
			}
		},
		userIDValidation,
		timeRangeValidation,
	)
}

// FetchRequest is validated input
type FetchRequest struct {
	UserID    types.UserID
	TimeRange types.TimeRange
}

// ============================================================================
// VALIDATION HELPERS
// ============================================================================

func validateUserID(userID types.UserID) either.Validation[[]validation.ValidationError, types.UserID] {
	errorMonoid := monoid.NewListMonoid[validation.ValidationError]()

	if userID.IsEmpty() {
		return either.Invalid[[]validation.ValidationError, types.UserID](
			errorMonoid,
			[]validation.ValidationError{{
				Field:   "user_id",
				Value:   string(userID),
				Message: "user_id is required",
				Code:    "required",
			}},
		)
	}

	return either.Valid[[]validation.ValidationError, types.UserID](errorMonoid, userID)
}

func validateTimeRange(tr types.TimeRange) either.Validation[[]validation.ValidationError, types.TimeRange] {
	errorMonoid := monoid.NewListMonoid[validation.ValidationError]()
	errors := []validation.ValidationError{}

	// Check if valid
	if !tr.IsValid() {
		errors = append(errors, validation.ValidationError{
			Field:   "time_range",
			Message: "start must be before end",
			Code:    "invalid_range",
		})
	}

	// Check if not too large (e.g., max 90 days)
	if tr.Duration() > 90*24*time.Hour {
		errors = append(errors, validation.ValidationError{
			Field:   "time_range",
			Message: "time range cannot exceed 90 days",
			Code:    "out_of_range",
		})
	}

	// Check if not in future
	if tr.End.After(time.Now()) {
		errors = append(errors, validation.ValidationError{
			Field:   "time_range.end",
			Message: "end time cannot be in the future",
			Code:    "invalid_date",
		})
	}

	if len(errors) > 0 {
		return either.Invalid[[]validation.ValidationError, types.TimeRange](errorMonoid, errors)
	}

	return either.Valid[[]validation.ValidationError, types.TimeRange](errorMonoid, tr)
}
