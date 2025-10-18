package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// MockClient simulates Jira API
type MockClient struct {
	config      types.JiraConfig
	latencyMS   int
	failureRate float64
}

// NewMockClient creates a mock Jira connector
func NewMockClient(config types.JiraConfig) *MockClient {
	return &MockClient{
		config:      config,
		latencyMS:   80 + rand.Intn(120),
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

// Fetch simulates fetching Jira events
func (c *MockClient) Fetch(
	ctx context.Context,
	userID types.UserID,
	tr types.TimeRange,
) ([]events.Event, error) {

	// Simulate API latency
	time.Sleep(time.Duration(c.latencyMS) * time.Millisecond)

	// Simulate random failures
	if rand.Float64() < c.failureRate {
		return nil, fmt.Errorf("jira api error: authentication failed")
	}

	return c.generateMockEvents(userID, tr), nil
}

// generateMockEvents creates realistic Jira events
func (c *MockClient) generateMockEvents(
	userID types.UserID,
	tr types.TimeRange,
) []events.Event {

	var mockEvents []events.Event

	// Realistic ticket summaries
	ticketSummaries := []string{
		"Implement user authentication with JWT",
		"Fix memory leak in connection pool",
		"Add pagination to REST API endpoints",
		"Upgrade React to version 18",
		"Database query optimization for reports",
		"Implement real-time notifications via WebSocket",
		"Add export to CSV functionality",
		"Fix race condition in worker threads",
		"Create automated backup system",
		"Implement OAuth2 integration",
		"Add monitoring and alerting dashboard",
		"Fix payment gateway timeout issues",
		"Refactor user service for better testability",
		"Implement rate limiting for API",
		"Add dark mode support to UI",
		"Security audit findings remediation",
		"Migrate from REST to GraphQL",
		"Performance improvements for search",
		"Add multi-factor authentication",
		"Implement data retention policies",
	}

	// Generate 5-15 issues created
	issueCount := 5 + rand.Intn(10)

	for i := 0; i < issueCount; i++ {
		timestamp := randomTimeInRange(tr)
		storyPoints := []int{1, 2, 3, 5, 8, 13}[rand.Intn(6)] // Fibonacci
		priority := []string{"Highest", "High", "Medium", "Low", "Lowest"}[rand.Intn(5)]

		payload := map[string]interface{}{
			"key":         fmt.Sprintf("PROJ-%d", 1234+rand.Intn(500)),
			"summary":     ticketSummaries[rand.Intn(len(ticketSummaries))],
			"type":        randomIssueType(),
			"status":      "To Do",
			"priority":    priority,
			"assignee":    map[string]string{"accountId": string(userID)},
			"storyPoints": storyPoints,
			"project":     randomProject(),
			"labels":      randomLabels(),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("jira-issue-created-%s-%d", userID, i)),
			UserID:    userID,
			Source:    types.PlatformJira,
			Type:      types.EventJiraIssueCreated,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomProject(),
				Size:    storyPoints,
			},
		})
	}

	// Generate 3-10 issues completed
	completedCount := 3 + rand.Intn(7)

	for i := 0; i < completedCount; i++ {
		timestamp := randomTimeInRange(tr)
		storyPoints := []int{1, 2, 3, 5, 8}[rand.Intn(5)]

		payload := map[string]interface{}{
			"key":         fmt.Sprintf("PROJ-%d", 1000+rand.Intn(234)),
			"summary":     ticketSummaries[rand.Intn(len(ticketSummaries))],
			"type":        randomIssueType(),
			"status":      "Done",
			"resolution":  "Fixed",
			"assignee":    map[string]string{"accountId": string(userID)},
			"storyPoints": storyPoints,
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("jira-issue-completed-%s-%d", userID, i)),
			UserID:    userID,
			Source:    types.PlatformJira,
			Type:      types.EventJiraIssueClosed,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomProject(),
				Size:    storyPoints,
			},
		})
	}

	// Generate 10-20 comments
	commentCount := 10 + rand.Intn(10)
	commentTexts := []string{
		"Updated the implementation based on code review feedback",
		"This is blocked by the infrastructure team",
		"Moving to next sprint due to dependencies",
		"QA testing complete, found 2 minor issues",
		"Performance testing shows 30% improvement",
		"Customer confirmed this resolves their issue",
		"Need clarification on the acceptance criteria",
		"PR is ready for review: github.com/org/repo/pull/123",
		"Deployed to staging for testing",
		"Security review approved with minor suggestions",
	}

	for i := 0; i < commentCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"body":   commentTexts[rand.Intn(len(commentTexts))],
			"author": map[string]string{"accountId": string(userID)},
			"issue":  fmt.Sprintf("PROJ-%d", 1000+rand.Intn(500)),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("jira-comment-%s-%d", userID, i)),
			UserID:    userID,
			Source:    types.PlatformJira,
			Type:      types.EventJiraComment,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomProject(),
			},
		})
	}

	// Generate 5-10 transitions
	transitionCount := 5 + rand.Intn(5)

	for i := 0; i < transitionCount; i++ {
		timestamp := randomTimeInRange(tr)
		transitions := [][]string{
			{"To Do", "In Progress"},
			{"In Progress", "Code Review"},
			{"Code Review", "QA Testing"},
			{"QA Testing", "Done"},
			{"In Progress", "Blocked"},
			{"Blocked", "In Progress"},
		}
		transition := transitions[rand.Intn(len(transitions))]

		payload := map[string]interface{}{
			"from":  transition[0],
			"to":    transition[1],
			"issue": fmt.Sprintf("PROJ-%d", 1000+rand.Intn(500)),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("jira-transition-%s-%d", userID, i)),
			UserID:    userID,
			Source:    types.PlatformJira,
			Type:      types.EventJiraTransition,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomProject(),
			},
		})
	}

	return mockEvents
}

// Helper functions
func randomProject() string {
	projects := []string{
		"BACKEND",
		"FRONTEND",
		"MOBILE",
		"INFRA",
		"DATA",
		"PLATFORM",
	}
	return projects[rand.Intn(len(projects))]
}

func randomIssueType() string {
	types := []string{"Story", "Task", "Bug", "Epic", "Spike", "Tech Debt"}
	return types[rand.Intn(len(types))]
}

func randomStatus() string {
	statuses := []string{"To Do", "In Progress", "Code Review", "QA Testing", "Done", "Blocked"}
	return statuses[rand.Intn(len(statuses))]
}

func randomLabels() []string {
	allLabels := []string{
		"backend", "frontend", "database", "api", "security",
		"performance", "bug-fix", "feature", "documentation", "testing",
	}
	numLabels := rand.Intn(3) + 1
	labels := make([]string, numLabels)
	for i := 0; i < numLabels; i++ {
		labels[i] = allLabels[rand.Intn(len(allLabels))]
	}
	return labels
}

func randomTimeInRange(tr types.TimeRange) time.Time {
	duration := tr.End.Sub(tr.Start)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	return tr.Start.Add(randomDuration)
}
