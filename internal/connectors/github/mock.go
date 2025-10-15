package github

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// MOCK GITHUB CONNECTOR
// ============================================================================

// MockClient simulates GitHub API
type MockClient struct {
	config      types.GitHubConfig
	latencyMS   int
	failureRate float64
}

// NewMockClient creates a mock GitHub connector
func NewMockClient(config types.GitHubConfig) *MockClient {
	return &MockClient{
		config:      config,
		latencyMS:   100 + rand.Intn(200), // 100-300ms latency (GitHub is slower)
		failureRate: 0.0,
	}
}

// WithLatency sets custom latency
func (c *MockClient) WithLatency(ms int) *MockClient {
	c.latencyMS = ms
	return c
}

// WithFailureRate sets failure probability
func (c *MockClient) WithFailureRate(rate float64) *MockClient {
	c.failureRate = rate
	return c
}

// Fetch simulates fetching GitHub events
func (c *MockClient) Fetch(
	ctx context.Context,
	userID types.UserID,
	tr types.TimeRange,
) ([]events.Event, error) {

	// Simulate API latency
	time.Sleep(time.Duration(c.latencyMS) * time.Millisecond)

	// Simulate random failures
	if rand.Float64() < c.failureRate {
		return nil, fmt.Errorf("github api error: secondary rate limit")
	}

	return c.generateMockEvents(userID, tr), nil
}

// generateMockEvents creates realistic GitHub events
func (c *MockClient) generateMockEvents(
	userID types.UserID,
	tr types.TimeRange,
) []events.Event {

	var mockEvents []events.Event

	// Generate 10-30 commits
	commitCount := 10 + rand.Intn(20)

	for i := 0; i < commitCount; i++ {
		timestamp := randomTimeInRange(tr)

		linesAdded := rand.Intn(500)
		linesDeleted := rand.Intn(200)

		payload := map[string]interface{}{
			"sha":     fmt.Sprintf("abc123%d", i),
			"message": fmt.Sprintf("feat: implement feature %d", i+1),
			"author": map[string]string{
				"login": string(userID),
				"email": fmt.Sprintf("%s@example.com", userID),
			},
			"stats": map[string]int{
				"additions": linesAdded,
				"deletions": linesDeleted,
				"total":     linesAdded + linesDeleted,
			},
			"repo": randomRepo(),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("github-commit-%d", i)),
			UserID:    userID,
			Source:    types.PlatformGitHub,
			Type:      types.EventGitHubCommit,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomRepo(),
				Size:    linesAdded + linesDeleted,
			},
		})
	}

	// Generate 3-8 pull requests
	prCount := 3 + rand.Intn(5)

	for i := 0; i < prCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"number": 1000 + i,
			"title":  fmt.Sprintf("Add feature %d", i+1),
			"state":  randomPRState(),
			"user": map[string]string{
				"login": string(userID),
			},
			"repo": randomRepo(),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("github-pr-%d", i)),
			UserID:    userID,
			Source:    types.PlatformGitHub,
			Type:      types.EventGitHubPR,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomRepo(),
			},
		})
	}

	// Generate 5-15 PR reviews
	reviewCount := 5 + rand.Intn(10)

	for i := 0; i < reviewCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"state": randomReviewState(),
			"user": map[string]string{
				"login": string(userID),
			},
			"body": fmt.Sprintf("LGTM! Great work on PR #%d", 1000+i),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("github-review-%d", i)),
			UserID:    userID,
			Source:    types.PlatformGitHub,
			Type:      types.EventGitHubPRReview,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomRepo(),
			},
		})
	}

	// Generate 2-5 issues
	issueCount := 2 + rand.Intn(3)

	for i := 0; i < issueCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"number": 500 + i,
			"title":  fmt.Sprintf("Bug: Fix issue %d", i+1),
			"state":  "open",
			"user": map[string]string{
				"login": string(userID),
			},
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("github-issue-%d", i)),
			UserID:    userID,
			Source:    types.PlatformGitHub,
			Type:      types.EventGitHubIssue,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomRepo(),
			},
		})
	}

	return mockEvents
}

// ============================================================================
// HELPERS
// ============================================================================

func randomRepo() string {
	repos := []string{
		"backend-api",
		"frontend-web",
		"mobile-app",
		"data-pipeline",
		"ml-models",
		"infrastructure",
	}
	return repos[rand.Intn(len(repos))]
}

func randomPRState() string {
	states := []string{"open", "open", "merged", "closed"} // More open/merged
	return states[rand.Intn(len(states))]
}

func randomReviewState() string {
	states := []string{"approved", "approved", "changes_requested", "commented"}
	return states[rand.Intn(len(states))]
}

func randomTimeInRange(tr types.TimeRange) time.Time {
	duration := tr.End.Sub(tr.Start)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	return tr.Start.Add(randomDuration)
}
