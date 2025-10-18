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
		latencyMS:   100 + rand.Intn(200),
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

	// Realistic commit messages
	commitMessages := []struct {
		message string
		adds    int
		dels    int
	}{
		{"feat: add user authentication middleware", 342, 18},
		{"fix: resolve memory leak in connection pool", 67, 45},
		{"refactor: extract payment processing to service layer", 523, 201},
		{"docs: update API documentation for v2 endpoints", 178, 23},
		{"test: add integration tests for checkout flow", 289, 0},
		{"perf: optimize database queries in report generation", 156, 234},
		{"chore: update dependencies to latest versions", 48, 52},
		{"feat: implement real-time notifications via WebSocket", 892, 67},
		{"fix: handle edge case in date parsing logic", 34, 12},
		{"style: apply consistent code formatting", 234, 198},
		{"build: update CI/CD pipeline configuration", 89, 45},
		{"feat: add GraphQL subscription support", 456, 23},
		{"fix: prevent SQL injection in search queries", 78, 34},
		{"refactor: migrate from class to functional components", 678, 890},
		{"test: increase code coverage to 85%", 345, 12},
		{"docs: add architecture decision records", 234, 0},
		{"feat: implement data export functionality", 567, 123},
		{"fix: correct timezone handling in scheduler", 89, 67},
		{"perf: add Redis caching layer", 345, 56},
		{"chore: remove deprecated API endpoints", 12, 456},
	}

	// Generate 10-30 commits
	commitCount := 10 + rand.Intn(20)

	for i := 0; i < commitCount; i++ {
		timestamp := randomTimeInRange(tr)
		commit := commitMessages[rand.Intn(len(commitMessages))]
		sha := generateSHA()

		payload := map[string]interface{}{
			"sha":     sha,
			"message": commit.message,
			"author": map[string]string{
				"login": string(userID),
				"email": fmt.Sprintf("%s@example.com", userID),
			},
			"stats": map[string]int{
				"additions": commit.adds,
				"deletions": commit.dels,
				"total":     commit.adds + commit.dels,
			},
			"repo":   randomRepo(),
			"branch": randomBranch(),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("github-commit-%s-%d", userID, i)),
			UserID:    userID,
			Source:    types.PlatformGitHub,
			Type:      types.EventGitHubCommit,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomRepo(),
				Size:    commit.adds + commit.dels,
			},
		})
	}

	// Realistic PR titles
	prTitles := []string{
		"Add caching layer for API responses",
		"Implement OAuth2 authentication flow",
		"Fix race condition in worker pool",
		"Upgrade React to version 18",
		"Add monitoring and alerting with Prometheus",
		"Refactor user service for better testability",
		"Implement data export functionality",
		"Add webhook notifications support",
		"Optimize image processing pipeline",
		"Add GraphQL subscriptions support",
		"Fix memory leak in event listeners",
		"Implement rate limiting for public APIs",
		"Add support for S3 file uploads",
		"Migrate database schema for multi-tenancy",
		"Implement circuit breaker pattern",
	}

	// Generate 3-8 pull requests
	prCount := 3 + rand.Intn(5)

	for i := 0; i < prCount; i++ {
		timestamp := randomTimeInRange(tr)
		prNumber := 1000 + rand.Intn(500)

		payload := map[string]interface{}{
			"number": prNumber,
			"title":  prTitles[rand.Intn(len(prTitles))],
			"state":  randomPRState(),
			"user": map[string]string{
				"login": string(userID),
			},
			"repo":      randomRepo(),
			"base":      "main",
			"head":      fmt.Sprintf("feature/PROJ-%d", 1000+rand.Intn(500)),
			"draft":     rand.Float32() < 0.2,
			"comments":  rand.Intn(15),
			"commits":   rand.Intn(10) + 1,
			"additions": rand.Intn(1000) + 50,
			"deletions": rand.Intn(500) + 10,
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("github-pr-%s-%d", userID, i)),
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
	reviewComments := []string{
		"LGTM! Great work on this implementation.",
		"Please add unit tests for the new functionality.",
		"Consider extracting this into a separate function for reusability.",
		"Nice optimization! This should improve performance significantly.",
		"Could we add error handling for this edge case?",
		"The logic looks good, but please update the documentation.",
		"This might cause a race condition. Consider using a mutex here.",
		"Great refactoring! Much cleaner than before.",
		"Please run the linter and fix the formatting issues.",
		"Can we add logging here for better debugging?",
	}

	for i := 0; i < reviewCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"state": randomReviewState(),
			"user": map[string]string{
				"login": string(userID),
			},
			"body":      reviewComments[rand.Intn(len(reviewComments))],
			"pr_number": 1000 + rand.Intn(500),
			"commit_id": generateSHA(),
			"path":      randomFilePath(),
			"line":      rand.Intn(500) + 1,
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("github-review-%s-%d", userID, i)),
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
	issueTitles := []string{
		"App crashes when uploading large files",
		"Search functionality returns incorrect results",
		"Memory leak in background job processor",
		"API returns 500 error intermittently",
		"UI freezes on mobile devices",
		"Data export fails for large datasets",
		"Authentication tokens expire prematurely",
		"Pagination broken on user list page",
	}

	for i := 0; i < issueCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"number": 500 + rand.Intn(200),
			"title":  issueTitles[rand.Intn(len(issueTitles))],
			"state":  randomIssueState(),
			"user": map[string]string{
				"login": string(userID),
			},
			"labels":    randomLabels(),
			"assignees": []string{string(userID)},
			"milestone": fmt.Sprintf("v2.%d.0", rand.Intn(5)),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("github-issue-%s-%d", userID, i)),
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

// Helper functions
func randomRepo() string {
	repos := []string{
		"backend-api",
		"frontend-web",
		"mobile-app",
		"data-pipeline",
		"ml-models",
		"infrastructure",
		"microservices",
		"analytics-engine",
	}
	return repos[rand.Intn(len(repos))]
}

func randomBranch() string {
	branches := []string{
		"main", "main", "main", // Main is most common
		"develop",
		"feature/new-feature",
		"bugfix/critical-fix",
		"hotfix/prod-issue",
	}
	return branches[rand.Intn(len(branches))]
}

func randomPRState() string {
	states := []string{"open", "open", "merged", "closed"}
	return states[rand.Intn(len(states))]
}

func randomReviewState() string {
	states := []string{"approved", "approved", "changes_requested", "commented"}
	return states[rand.Intn(len(states))]
}

func randomIssueState() string {
	states := []string{"open", "open", "closed"}
	return states[rand.Intn(len(states))]
}

func randomFilePath() string {
	paths := []string{
		"src/services/user.service.ts",
		"src/controllers/auth.controller.ts",
		"src/models/product.model.ts",
		"src/utils/validation.ts",
		"src/middleware/auth.middleware.ts",
		"tests/integration/api.test.ts",
		"src/components/Dashboard.tsx",
		"src/hooks/useAuth.ts",
	}
	return paths[rand.Intn(len(paths))]
}

func randomLabels() []string {
	allLabels := []string{
		"bug", "enhancement", "documentation", "good first issue",
		"help wanted", "performance", "security", "testing",
	}
	numLabels := rand.Intn(3) + 1
	labels := make([]string, numLabels)
	for i := 0; i < numLabels; i++ {
		labels[i] = allLabels[rand.Intn(len(allLabels))]
	}
	return labels
}

func generateSHA() string {
	chars := "0123456789abcdef"
	sha := make([]byte, 40)
	for i := range sha {
		sha[i] = chars[rand.Intn(len(chars))]
	}
	return string(sha)
}

func randomTimeInRange(tr types.TimeRange) time.Time {
	duration := tr.End.Sub(tr.Start)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	return tr.Start.Add(randomDuration)
}
