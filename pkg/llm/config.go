// ============================================================================
// pkg/llm/config.go - Configuration and Builders
// ============================================================================

package llm

import (
	"database/sql"
)

// ============================================================================
// CONFIGURATION
// ============================================================================

// Config for LLM layer
type Config struct {
	ClaudeAPIKey   ClaudeAPIKey
	ClaudeModel    ClaudeModel
	ClaudeBaseURL  ClaudeBaseURL
	MaxRetries     MaxRetries
	Timeout        TimeoutDuration
	MaxConcurrency MaxConcurrency
	EnableCaching  bool
	EnableLogging  bool
	RateLimitRPM   RateLimitRPM
	RateLimitTPM   RateLimitTPM
}

// ============================================================================
// CONSTRUCTORS
// ============================================================================

// NewClaudeClient creates a new Claude client
func NewClaudeClient(config Config) *ClaudeClient {
	return nil
}

// NewSummaryGenerator creates a new summary generator
func NewSummaryGenerator(client *ClaudeClient) *SummaryGenerator {
	return nil
}

// NewSummaryRepository creates a new summary repository
func NewSummaryRepository(db *sql.DB) *SummaryRepository {
	return nil
}

// ============================================================================
// CLIENT BUILDER (Builder Pattern)
// ============================================================================

// ClientBuilder builds Claude clients with fluent API
type ClientBuilder struct {
	config Config
}

// NewClientBuilder creates a new client builder
func NewClientBuilder() *ClientBuilder {
	return nil
}

// WithAPIKey sets the API key
func (cb *ClientBuilder) WithAPIKey(key ClaudeAPIKey) *ClientBuilder {
	return cb
}

// WithModel sets the model
func (cb *ClientBuilder) WithModel(model ClaudeModel) *ClientBuilder {
	return cb
}

// WithTimeout sets the timeout
func (cb *ClientBuilder) WithTimeout(timeout TimeoutDuration) *ClientBuilder {
	return cb
}

// WithRetries sets the max retries
func (cb *ClientBuilder) WithRetries(retries MaxRetries) *ClientBuilder {
	return cb
}

// WithRateLimit sets the rate limit
func (cb *ClientBuilder) WithRateLimit(rpm RateLimitRPM, tpm RateLimitTPM) *ClientBuilder {
	return cb
}

// Build constructs the client
func (cb *ClientBuilder) Build() *ClaudeClient {
	return nil
}
