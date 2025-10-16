package llm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/effect"
	"github.com/vinodhalaharvi/purekernels/pkg/result"
)

// ============================================================================
// CLAUDE API REQUEST/RESPONSE TYPES
// ============================================================================

// claudeRequest is the Claude API request format
type claudeRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature,omitempty"`
	Messages    []claudeMessage `json:"messages"`
	System      string          `json:"system,omitempty"`
}

// claudeMessage represents a message in the conversation
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse is the Claude API response format
type claudeResponse struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"`
	Role         string               `json:"role"`
	Content      []claudeContentBlock `json:"content"`
	Model        string               `json:"model"`
	StopReason   string               `json:"stop_reason"`
	StopSequence string               `json:"stop_sequence"`
	Usage        claudeUsage          `json:"usage"`
}

// claudeContentBlock represents a content block in the response
type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// claudeUsage tracks token usage
type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ============================================================================
// CLAUDE CLIENT METHODS (Impure - IO Boundary)
// ============================================================================

// Generate calls Claude API (impure)
func (c *ClaudeClient) Generate(ctx context.Context, prompt Prompt) (LLMResponse, error) {
	// Build request
	reqBody := claudeRequest{
		Model:       c.Model,
		MaxTokens:   prompt.MaxTokens,
		Temperature: prompt.Temperature,
		System:      prompt.SystemMessage,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: prompt.UserMessage,
			},
		},
	}

	// Marshal request
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: c.Timeout,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return LLMResponse{}, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return LLMResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text content
	content := ""
	if len(claudeResp.Content) > 0 {
		content = claudeResp.Content[0].Text
	}

	// Build LLMResponse
	return LLMResponse{
		Content:      content,
		StopReason:   claudeResp.StopReason,
		ModelVersion: claudeResp.Model,
		Usage: TokenUsage{
			InputTokens:  claudeResp.Usage.InputTokens,
			OutputTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:  claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
		Timestamp: time.Now(),
	}, nil
}

// Effect wraps Generate in an LLMEffect
func (c *ClaudeClient) Effect(prompt Prompt) LLMEffect[LLMResponse] {
	return LiftIO(func(ctx context.Context) (LLMResponse, error) {
		return c.Generate(ctx, prompt)
	})
}

// GenerateBatch calls Claude API for multiple prompts sequentially
func (c *ClaudeClient) GenerateBatch(ctx context.Context, prompts []Prompt) ([]LLMResponse, error) {
	responses := make([]LLMResponse, 0, len(prompts))

	for _, prompt := range prompts {
		response, err := c.Generate(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate response for prompt: %w", err)
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// ============================================================================
// LLM EFFECT CONSTRUCTORS
// ============================================================================

// Pure wraps a pure value in an effect

// ============================================================================
// CONCURRENT GENERATION METHODS
// ============================================================================

// GenerateAll creates concurrent effects for all prompts
func (cg *ConcurrentLLM) GenerateAll(prompts []Prompt) []LLMEffect[LLMResponse] {
	effects := make([]LLMEffect[LLMResponse], len(prompts))

	for i, prompt := range prompts {
		effects[i] = cg.Client.Effect(prompt)
	}

	return effects
}

// GenerateConcurrent executes all prompts in parallel and combines results
func (cg *ConcurrentLLM) GenerateConcurrent(
	ctx context.Context,
	prompts []Prompt,
) effect.Writer[[]string, result.Result[[]LLMResponse]] {
	// Create effects for all prompts
	effects := cg.GenerateAll(prompts)

	// Execute all effects in parallel
	responses := make([]LLMResponse, len(effects))
	allLogs := []string{}
	errors := []error{}

	// Use goroutines for parallel execution
	type indexedResult struct {
		index    int
		response LLMResponse
		logs     []string
		err      error
	}

	resultsChan := make(chan indexedResult, len(effects))

	for i, eff := range effects {
		go func(idx int, effect LLMEffect[LLMResponse]) {
			writer := effect.Run(ctx)
			res, logs := writer.Run()

			if res.IsOk() {
				resultsChan <- indexedResult{
					index:    idx,
					response: res.Unwrap(),
					logs:     logs,
					err:      nil,
				}
			} else {
				resultsChan <- indexedResult{
					index: idx,
					logs:  logs,
					err:   res.Error(),
				}
			}
		}(i, eff)
	}

	// Collect results
	for range effects {
		res := <-resultsChan
		allLogs = append(allLogs, res.logs...)

		if res.err != nil {
			errors = append(errors, res.err)
		} else {
			responses[res.index] = res.response
		}
	}

	// Return combined result
	if len(errors) > 0 {
		return effect.NewWriter(
			result.Err[[]LLMResponse](fmt.Errorf("concurrent generation failed: %d errors", len(errors))),
			allLogs,
		)
	}

	return effect.NewWriter(result.Ok(responses), allLogs)
}

// ============================================================================
// CACHING METHODS
// ============================================================================

// GenerateWithCache checks cache first, then calls API if needed
func (cg *CachedLLM) GenerateWithCache(
	prompt Prompt,
) effect.State[map[string]LLMResponse, result.Result[LLMResponse]] {
	return effect.NewState(func(cache map[string]LLMResponse) (result.Result[LLMResponse], map[string]LLMResponse) {
		// Check cache
		hash := HashPrompt(prompt)
		if cached, ok := cache[hash]; ok {
			// Cache hit
			return result.Ok(cached), cache
		}

		// Cache miss - call API (this breaks purity, but it's at the boundary)
		ctx := context.Background()
		response, err := cg.Client.Generate(ctx, prompt)

		if err != nil {
			return result.Err[LLMResponse](err), cache
		}

		// Update cache
		newCache := make(map[string]LLMResponse, len(cache)+1)
		for k, v := range cache {
			newCache[k] = v
		}
		newCache[hash] = response

		return result.Ok(response), newCache
	})
}

// HashPrompt creates a cache key from a prompt
func HashPrompt(prompt Prompt) string {
	// Create deterministic hash from prompt content
	data := fmt.Sprintf("%s|%s|%d|%.2f",
		prompt.SystemMessage,
		prompt.UserMessage,
		prompt.MaxTokens,
		prompt.Temperature,
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func Pure[A any](value A) LLMEffect[A] {
	return LLMEffect[A]{
		Run: func(ctx context.Context) effect.Writer[[]string, result.Result[A]] {
			// Swap parameters: value first, log second
			return effect.NewWriter(result.Ok(value), []string{})
		},
	}
}

// LiftIO lifts an impure function into an effect
func LiftIO[A any](fn func(context.Context) (A, error)) LLMEffect[A] {
	return LLMEffect[A]{
		Run: func(ctx context.Context) effect.Writer[[]string, result.Result[A]] {
			// Execute the impure function
			value, err := fn(ctx)

			// Log the execution
			logs := []string{
				fmt.Sprintf("llm_call_started: %s", time.Now().Format(time.RFC3339)),
			}

			// Create result
			var res result.Result[A]
			if err != nil {
				logs = append(logs, fmt.Sprintf("llm_call_failed: %v", err))
				res = result.Err[A](err)
			} else {
				logs = append(logs, "llm_call_succeeded")
				res = result.Ok(value)
			}

			// Swap parameters: value first, log second
			return effect.NewWriter(res, logs)
		},
	}
}
