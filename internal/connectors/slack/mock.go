package slack

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
// MOCK SLACK CONNECTOR
// ============================================================================

// MockClient simulates Slack API
type MockClient struct {
	config      types.SlackConfig
	latencyMS   int     // Simulated API latency
	shouldFail  bool    // Simulate failures
	failureRate float64 // 0.0 - 1.0
}

// NewMockClient creates a mock Slack connector
func NewMockClient(config types.SlackConfig) *MockClient {
	return &MockClient{
		config:      config,
		latencyMS:   50 + rand.Intn(100), // 50-150ms latency
		shouldFail:  false,
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

// Fetch simulates fetching Slack events
func (c *MockClient) Fetch(
	ctx context.Context,
	userID types.UserID,
	tr types.TimeRange,
) ([]events.Event, error) {

	// Simulate API latency
	time.Sleep(time.Duration(c.latencyMS) * time.Millisecond)

	// Simulate random failures
	if rand.Float64() < c.failureRate {
		return nil, fmt.Errorf("slack api error: rate limit exceeded")
	}

	// Generate mock events
	return c.generateMockEvents(userID, tr), nil
}

// generateMockEvents creates realistic Slack events
func (c *MockClient) generateMockEvents(
	userID types.UserID,
	tr types.TimeRange,
) []events.Event {

	var mockEvents []events.Event

	// Generate 20-50 messages
	messageCount := 20 + rand.Intn(30)

	for i := 0; i < messageCount; i++ {
		timestamp := randomTimeInRange(tr)

		// Create mock message payload
		payload := map[string]interface{}{
			"text":    fmt.Sprintf("Mock Slack message %d", i+1),
			"channel": randomChannel(),
			"ts":      fmt.Sprintf("%d.%06d", timestamp.Unix(), rand.Intn(999999)),
			"user":    string(userID),
			"team":    c.config.WorkspaceID,
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("slack-msg-%d", i)),
			UserID:    userID,
			Source:    types.PlatformSlack,
			Type:      types.EventSlackMessage,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomChannel(),
				Size:    20 + rand.Intn(200), // Message length
			},
		})
	}

	// Generate 5-15 reactions
	reactionCount := 5 + rand.Intn(10)

	for i := 0; i < reactionCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"reaction": randomEmoji(),
			"user":     string(userID),
			"item": map[string]string{
				"type":    "message",
				"channel": randomChannel(),
			},
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("slack-reaction-%d", i)),
			UserID:    userID,
			Source:    types.PlatformSlack,
			Type:      types.EventSlackReaction,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomChannel(),
			},
		})
	}

	// Generate 2-5 file uploads
	fileCount := 2 + rand.Intn(3)

	for i := 0; i < fileCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"name":     fmt.Sprintf("document-%d.pdf", i+1),
			"mimetype": "application/pdf",
			"size":     1024 * (100 + rand.Intn(900)), // 100KB - 1MB
			"user":     string(userID),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("slack-file-%d", i)),
			UserID:    userID,
			Source:    types.PlatformSlack,
			Type:      types.EventSlackFileUpload,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: randomChannel(),
				Size:    1024 * (100 + rand.Intn(900)),
			},
		})
	}

	return mockEvents
}

// ============================================================================
// HELPERS
// ============================================================================

func randomChannel() string {
	channels := []string{
		"general",
		"engineering",
		"product",
		"design",
		"marketing",
		"random",
		"watercooler",
	}
	return channels[rand.Intn(len(channels))]
}

func randomEmoji() string {
	emojis := []string{
		"thumbsup",
		"heart",
		"rocket",
		"eyes",
		"fire",
		"100",
		"party",
		"clap",
	}
	return emojis[rand.Intn(len(emojis))]
}

func randomTimeInRange(tr types.TimeRange) time.Time {
	duration := tr.End.Sub(tr.Start)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	return tr.Start.Add(randomDuration)
}
