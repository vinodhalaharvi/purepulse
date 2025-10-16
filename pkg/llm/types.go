package llm

import (
	"context"
	"database/sql"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/effect"
	"github.com/vinodhalaharvi/purekernels/pkg/either"
	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purekernels/pkg/result"
	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// PROMPT (Pure - Monoid)
// ============================================================================

// Prompt represents a structured LLM prompt
type Prompt struct {
	SystemMessage string
	UserMessage   string
	Context       map[string]interface{}
	MaxTokens     int
	Temperature   float64
}

// PromptMonoid is the monoid instance for Prompt
type PromptMonoid struct{}

// PromptBuilder builds prompts from input data
type PromptBuilder[A any] struct {
	Build func(A) Prompt
}

// ============================================================================
// LLM RESPONSE
// ============================================================================

// LLMResponse represents Claude's response
type LLMResponse struct {
	Content      string
	StopReason   string
	Usage        TokenUsage
	ModelVersion string
	Timestamp    time.Time
}

// TokenUsage tracks API token consumption
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// ============================================================================
// VALIDATION
// ============================================================================

// LLMValidation uses either.Validation from purekernels
type LLMValidation = either.Validation[[]ValidationError, LLMResponse]

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrorsMonoid accumulates validation errors
type ValidationErrorsMonoid struct{}

// ============================================================================
// SUMMARY (Domain Model)
// ============================================================================

// Summary is the structured output from LLM
type Summary struct {
	ID          string
	UserID      types.UserID
	Type        SummaryType
	Content     string
	Highlights  []string
	Insights    []string
	Metrics     SummaryMetrics
	TimeRange   types.TimeRange
	Platforms   []types.Platform
	GeneratedAt time.Time
	ModelUsed   string
	TokenUsage  TokenUsage
}

// SummaryType categorizes summaries
type SummaryType string

const (
	SummaryTypeDaily        SummaryType = "daily"
	SummaryTypeWeekly       SummaryType = "weekly"
	SummaryTypeProductivity SummaryType = "productivity"
	SummaryTypeInsights     SummaryType = "insights"
)

// SummaryMetrics contains computed metrics
type SummaryMetrics struct {
	TotalEvents        int
	PlatformsActive    int
	ProductivityScore  float64
	FocusScore         float64
	CollaborationScore float64
}

// SummaryMonoid combines summaries
type SummaryMonoid struct{}

// ============================================================================
// LLM EFFECT
// ============================================================================

// LLMEffect wraps an impure LLM call with logging
type LLMEffect[A any] struct {
	Run func(context.Context) effect.Writer[[]string, result.Result[A]]
}

// ============================================================================
// CLAUDE CLIENT
// ============================================================================

// ClaudeClient is the impure HTTP client
type ClaudeClient struct {
	APIKey     string
	Model      string
	BaseURL    string
	MaxRetries int
	Timeout    time.Duration
}

// ============================================================================
// CONCURRENT GENERATION
// ============================================================================

// ConcurrentLLM generates multiple responses in parallel
type ConcurrentLLM struct {
	Client *ClaudeClient
	Monoid monoid.Monoid[[]LLMResponse]
}

// LLMResponseListMonoid combines lists of responses
type LLMResponseListMonoid struct{}

// ============================================================================
// CACHING LAYER
// ============================================================================

// CachedLLM wraps generation with caching using State monad
type CachedLLM struct {
	Client *ClaudeClient
}

// ============================================================================
// SUMMARY GENERATOR
// ============================================================================

// SummaryGenerator orchestrates the full pipeline
type SummaryGenerator struct {
	Client         *ClaudeClient
	PromptBuilder  PromptBuilder[[]events.Event]
	ResponseParser func(LLMResponse) result.Result[Summary]
	Validator      func(LLMResponse) LLMValidation
}

// ============================================================================
// PIPELINE COMPOSITION
// ============================================================================

// ============================================================================
// DATABASE PERSISTENCE
// ============================================================================

// SummaryRepository handles summary persistence
type SummaryRepository struct {
	DB *sql.DB
}

// ============================================================================
// PIPELINE COMPOSITION
// ============================================================================

// Pipeline represents a composable LLM pipeline
type Pipeline[A, B any] struct {
	Run func(A) result.Result[B]
}
