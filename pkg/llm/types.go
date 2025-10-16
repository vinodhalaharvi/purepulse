// ============================================================================
// pkg/llm/types.go - Type Aliases and Core Types
// ============================================================================

package llm

import (
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// PRIMITIVE TYPE ALIASES (Type Safety)
// ============================================================================

// Client configuration types
type ClaudeAPIKey string
type ClaudeModel string
type ClaudeBaseURL string
type MaxRetries int
type TimeoutDuration time.Duration
type MaxConcurrency int
type RateLimitRPM int // Requests per minute
type RateLimitTPM int // Tokens per minute

// Prompt types
type SystemMessage string
type UserMessage string
type MaxTokens int
type Temperature float64

// Response types
type ResponseContent string
type StopReason string
type ModelVersion string

// Token usage types
type InputTokens int
type OutputTokens int
type TotalTokens int
type TokenCost float64 // USD

// Validation types
type ValidationField string
type ValidationMessage string
type ValidationCode string

// Summary types
type SummaryID string
type SummaryContent string
type Highlight string
type Insight string

// Metrics types
type EventCount int
type PlatformCount int
type ProductivityScore float64
type FocusScore float64
type CollaborationScore float64
type DeepWorkHours float64
type MeetingHours float64
type CodeCommitCount int
type PullRequestCount int
type IssueCount int
type MessageCount int

// Cache types
type CacheKey string

// ============================================================================
// CORE DOMAIN TYPES
// ============================================================================

// Prompt represents a structured LLM prompt (Monoid)
type Prompt struct {
	SystemMessage SystemMessage
	UserMessage   UserMessage
	Context       map[string]interface{}
	MaxTokens     MaxTokens
	Temperature   Temperature
}

// LLMResponse represents a response from Claude
type LLMResponse struct {
	Content      ResponseContent
	StopReason   StopReason
	Usage        TokenUsage
	ModelVersion ModelVersion
	Timestamp    time.Time
	Metadata     map[string]interface{}
}

// TokenUsage tracks token consumption and cost
type TokenUsage struct {
	InputTokens  InputTokens
	OutputTokens OutputTokens
	TotalTokens  TotalTokens
	Cost         TokenCost
}

// ValidationError for LLM responses
type ValidationError struct {
	Field   ValidationField
	Message ValidationMessage
	Code    ValidationCode
}

// LLMValidation wraps a response with validation errors (Validation Applicative)
type LLMValidation struct {
	Response LLMResponse
	Errors   []ValidationError
}

// Summary is the structured output from LLM
type Summary struct {
	ID          SummaryID
	UserID      types.UserID
	Type        SummaryType
	Content     SummaryContent
	Highlights  []Highlight
	Insights    []Insight
	Metrics     SummaryMetrics
	TimeRange   types.TimeRange
	Platforms   []types.Platform
	GeneratedAt time.Time
	ModelUsed   ModelVersion
	TokenUsage  TokenUsage
	Metadata    map[string]interface{}
}

// SummaryType enum
type SummaryType string

const (
	SummaryTypeDaily        SummaryType = "daily"
	SummaryTypeWeekly       SummaryType = "weekly"
	SummaryTypeMonthly      SummaryType = "monthly"
	SummaryTypeProductivity SummaryType = "productivity"
	SummaryTypeInsights     SummaryType = "insights"
	SummaryTypeHighlights   SummaryType = "highlights"
)

// SummaryMetrics contains computed metrics
type SummaryMetrics struct {
	TotalEvents        EventCount
	PlatformsActive    PlatformCount
	ProductivityScore  ProductivityScore
	FocusScore         FocusScore
	CollaborationScore CollaborationScore
	ResponseTimeAvg    time.Duration
	DeepWorkHours      DeepWorkHours
	MeetingHours       MeetingHours
	CodeCommits        CodeCommitCount
	PullRequests       PullRequestCount
	IssuesClosed       IssueCount
	MessagesExchanged  MessageCount
}

// PromptBuilder builds prompts from events (Functor)
type PromptBuilder[A any] struct {
	Build func(A) Prompt
}

// RateLimiter controls API call rate
type RateLimiter struct {
	RequestsPerMinute RateLimitRPM
	TokensPerMinute   RateLimitTPM
}
