package validation

import (
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// VALIDATION
// ============================================================================

// ValidationError for input validation failures
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Code    string `json:"code"` // "required", "invalid_format", "out_of_range"
}

// Error implements error interface
func (ve ValidationError) Error() string {
	return ve.Message
}

// ============================================================================
// EXECUTION PLANNING
// ============================================================================

// ExecutionPlan describes fetch strategy
type ExecutionPlan struct {
	UserID    types.UserID     `json:"user_id"`
	TimeRange types.TimeRange  `json:"time_range"`
	Sources   []types.Platform `json:"sources"`
	Steps     []PlanStep       `json:"steps"`
	Estimated CostEstimate     `json:"estimated"`
	CreatedAt time.Time        `json:"created_at"`
}

// PlanStep represents individual fetch step
type PlanStep struct {
	Source       types.Platform `json:"source"`
	Operation    string         `json:"operation"` // "fetch_messages", "fetch_commits", etc.
	EstimatedMS  int            `json:"estimated_ms"`
	EstimatedAPI int            `json:"estimated_api_calls"`
	CacheCheck   bool           `json:"cache_check"`
}

// CostEstimate for resource planning
type CostEstimate struct {
	TotalAPICalls int     `json:"total_api_calls"`
	TotalTimeMS   int     `json:"total_time_ms"`
	CacheHitRate  float64 `json:"cache_hit_rate"`
	RateLimitRisk bool    `json:"rate_limit_risk"`
}
