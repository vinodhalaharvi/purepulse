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

// MockClient simulates Slack API
type MockClient struct {
	config      types.SlackConfig
	latencyMS   int
	shouldFail  bool
	failureRate float64
}

// NewMockClient creates a mock Slack connector
func NewMockClient(config types.SlackConfig) *MockClient {
	return &MockClient{
		config:      config,
		latencyMS:   50 + rand.Intn(100),
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

	// Realistic message templates
	messages := []string{
		"Just pushed the fix for the authentication issue to staging",
		"Can someone review PR #1234 for the payment gateway?",
		"Deployed v2.3.4 to production - all systems green",
		"Found a critical bug in the checkout flow, working on a fix",
		"Sprint planning notes uploaded to Confluence",
		"Database migration completed successfully, no downtime",
		"API rate limiting implemented as discussed in the RFC",
		"Customer reported issue with password reset flow - investigating",
		"Performance improvements deployed - seeing 40% reduction in latency",
		"Updated the OpenAPI documentation for the new endpoints",
		"Staging environment is down, investigating with DevOps",
		"Security patch applied to all production servers",
		"Code review feedback addressed, ready for merge",
		"New feature flag enabled for 10% of users",
		"Incident resolved - post-mortem scheduled for tomorrow",
	}

	// Generate 20-50 messages
	messageCount := 20 + rand.Intn(30)

	for i := 0; i < messageCount; i++ {
		timestamp := randomTimeInRange(tr)
		channel := randomChannel()
		message := messages[rand.Intn(len(messages))]

		payload := map[string]interface{}{
			"text":    message,
			"channel": channel,
			"ts":      fmt.Sprintf("%d.%06d", timestamp.Unix(), rand.Intn(999999)),
			"user":    string(userID),
			"team":    c.config.WorkspaceID,
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("slack-msg-%s-%d", userID, i)),
			UserID:    userID,
			Source:    types.PlatformSlack,
			Type:      types.EventSlackMessage,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: channel,
				Size:    len(message),
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
				"type":    "slack_message",
				"channel": randomChannel(),
			},
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("slack-reaction-%s-%d", userID, i)),
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
	fileNames := []string{
		"architecture-diagram.pdf",
		"sprint-report.xlsx",
		"api-documentation.md",
		"test-results.json",
		"deployment-guide.pdf",
		"performance-metrics.csv",
		"security-audit.pdf",
	}

	for i := 0; i < fileCount; i++ {
		timestamp := randomTimeInRange(tr)
		fileName := fileNames[rand.Intn(len(fileNames))]

		payload := map[string]interface{}{
			"name":     fileName,
			"mimetype": getMimeType(fileName),
			"size":     1024 * (100 + rand.Intn(900)),
			"user":     string(userID),
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("slack-file-%s-%d", userID, i)),
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

// Helper functions
func randomChannel() string {
	channels := []string{
		"#engineering",
		"#product",
		"#design",
		"#standup",
		"#incidents",
		"#releases",
		"#backend",
		"#frontend",
		"#devops",
		"#random",
	}
	return channels[rand.Intn(len(channels))]
}

func randomEmoji() string {
	emojis := []string{
		"thumbsup",
		"white_check_mark",
		"rocket",
		"eyes",
		"fire",
		"100",
		"ship",
		"tada",
	}
	return emojis[rand.Intn(len(emojis))]
}

func getMimeType(filename string) string {
	if len(filename) > 4 {
		ext := filename[len(filename)-4:]
		switch ext {
		case ".pdf":
			return "application/pdf"
		case ".csv":
			return "text/csv"
		case "xlsx":
			return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		case "json":
			return "application/json"
		case ".md":
			return "text/markdown"
		}
	}
	return "application/octet-stream"
}

func randomTimeInRange(tr types.TimeRange) time.Time {
	duration := tr.End.Sub(tr.Start)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	return tr.Start.Add(randomDuration)
}
