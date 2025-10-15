package http

import (
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// HTTP ERRORS
// ============================================================================

// HTTPError for HTTP-level errors
type HTTPError struct {
	StatusCode types.StatusCode `json:"status_code"`
	Message    string           `json:"message"`
	Code       string           `json:"code"` // "invalid_request", "not_found", "internal_error"
	Details    []string         `json:"details,omitempty"`
	Timestamp  time.Time        `json:"timestamp"`
}

// Error implements error interface
func (he HTTPError) Error() string {
	return he.Message
}

// EncodingError for JSON encoding failures
type EncodingError struct {
	Field   string `json:"field"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Error implements error interface
func (ee EncodingError) Error() string {
	return ee.Message
}

// DecodingError for JSON decoding failures
type DecodingError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implements error interface
func (de DecodingError) Error() string {
	return de.Message
}
