// ============================================================================
// pkg/llm/client.go - Claude API Client (Impure IO Boundary)
// ============================================================================

package llm

import (
	"context"
)

// ============================================================================
// CLAUDE CLIENT
// ============================================================================

// ClaudeClient is the impure LLM client
type ClaudeClient struct {
	APIKey     ClaudeAPIKey
	Model      ClaudeModel
	BaseURL    ClaudeBaseURL
	MaxRetries MaxRetries
	Timeout    TimeoutDuration
	RateLimit  *RateLimiter
}

// ============================================================================
// CLIENT OPERATIONS (Impure IO)
// ============================================================================

// Generate calls Claude API (impure IO)
func (c *ClaudeClient) Generate(ctx context.Context, prompt Prompt) (LLMResponse, error) {
	return LLMResponse{}, nil
}

// Effect wraps Generate in LLMEffect
func (c *ClaudeClient) Effect(prompt Prompt) LLMEffect[LLMResponse] {
	return LLMEffect[LLMResponse]{}
}

// GenerateBatch calls Claude API in parallel for multiple prompts
func (c *ClaudeClient) GenerateBatch(ctx context.Context, prompts []Prompt) ([]LLMResponse, error) {
	return nil, nil
}

// EffectBatch wraps GenerateBatch in LLMEffect
func (c *ClaudeClient) EffectBatch(prompts []Prompt) LLMEffect[[]LLMResponse] {
	return LLMEffect[[]LLMResponse]{}
}

// Stream for streaming responses (optional, for future)
func (c *ClaudeClient) Stream(ctx context.Context, prompt Prompt) (<-chan string, error) {
	return nil, nil
}

// Ping checks API health
func (c *ClaudeClient) Ping(ctx context.Context) error {
	return nil
}
