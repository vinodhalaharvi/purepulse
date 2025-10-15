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

// ============================================================================
// MOCK JIRA CONNECTOR
// ============================================================================

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
		latencyMS:   80 + rand.Intn(120), // 80-200ms latency
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

	// Generate 5-15 issues created
	issueCount := 5 + rand.Intn(10)

	for i := 0; i < issueCount; i++ {
		timestamp := randomTimeInRange(tr)

		storyPoints := rand.Intn(8) + 1

		payload := map[string]interface{}{
			"key":     fmt.Sprintf("PROJ-%d", 1000+i),
			"summary": fmt.Sprintf("Implement feature %d", i+1),
			"type":    randomIssueType(),
			"status":  "To Do",
			"assignee": map[string]string{
				"accountId": string(userID),
			},
			"storyPoints": storyPoints,
			"project":     randomProject(),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("jira-issue-created-%d", i)),
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

		storyPoints := rand.Intn(8) + 1

		payload := map[string]interface{}{
			"key":     fmt.Sprintf("PROJ-%d", 500+i),
			"summary": fmt.Sprintf("Complete task %d", i+1),
			"type":    randomIssueType(),
			"status":  "Done",
			"assignee": map[string]string{
				"accountId": string(userID),
			},
			"storyPoints": storyPoints,
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("jira-issue-completed-%d", i)),
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

	for i := 0; i < commentCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"body": fmt.Sprintf("Comment on issue %d", i+1),
			"author": map[string]string{
				"accountId": string(userID),
			},
			"issue": fmt.Sprintf("PROJ-%d", rand.Intn(500)),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("jira-comment-%d", i)),
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

		payload := map[string]interface{}{
			"from":  randomStatus(),
			"to":    randomStatus(),
			"issue": fmt.Sprintf("PROJ-%d", rand.Intn(500)),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("jira-transition-%d", i)),
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

// ============================================================================
// HELPERS
// ============================================================================

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
	types := []string{"Story", "Task", "Bug", "Epic"}
	return types[rand.Intn(len(types))]
}

func randomStatus() string {
	statuses := []string{"To Do", "In Progress", "In Review", "Done"}
	return statuses[rand.Intn(len(statuses))]
}

func randomTimeInRange(tr types.TimeRange) time.Time {
	duration := tr.End.Sub(tr.Start)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	return tr.Start.Add(randomDuration)
}
