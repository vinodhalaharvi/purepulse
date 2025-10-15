package zoom

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
// MOCK ZOOM CONNECTOR
// ============================================================================

// MockClient simulates Zoom API
type MockClient struct {
	config      types.ZoomConfig
	latencyMS   int
	failureRate float64
}

// NewMockClient creates a mock Zoom connector
func NewMockClient(config types.ZoomConfig) *MockClient {
	return &MockClient{
		config:      config,
		latencyMS:   70 + rand.Intn(130), // 70-200ms latency
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

// Fetch simulates fetching Zoom events
func (c *MockClient) Fetch(
	ctx context.Context,
	userID types.UserID,
	tr types.TimeRange,
) ([]events.Event, error) {

	// Simulate API latency
	time.Sleep(time.Duration(c.latencyMS) * time.Millisecond)

	// Simulate random failures
	if rand.Float64() < c.failureRate {
		return nil, fmt.Errorf("zoom api error: invalid access token")
	}

	return c.generateMockEvents(userID, tr), nil
}

// generateMockEvents creates realistic Zoom events
func (c *MockClient) generateMockEvents(
	userID types.UserID,
	tr types.TimeRange,
) []events.Event {

	var mockEvents []events.Event

	// Generate 5-15 meetings
	meetingCount := 5 + rand.Intn(10)

	for i := 0; i < meetingCount; i++ {
		timestamp := randomTimeInRange(tr)
		durationMinutes := 15 + rand.Intn(90) // 15-105 minutes
		participantCount := 2 + rand.Intn(8)  // 2-10 participants

		payload := map[string]interface{}{
			"id":       fmt.Sprintf("zoom-meeting-%d", i),
			"topic":    randomMeetingTopic(),
			"duration": durationMinutes,
			"host": map[string]string{
				"id":    string(userID),
				"email": fmt.Sprintf("%s@example.com", userID),
			},
			"participants": participantCount,
			"type":         randomMeetingType(),
		}

		payloadJSON, _ := json.Marshal(payload)

		durationSeconds := durationMinutes * 60

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("zoom-meeting-%d", i)),
			UserID:    userID,
			Source:    types.PlatformZoom,
			Type:      types.EventZoomMeeting,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:          string(userID),
				DurationSeconds: &durationSeconds,
				Participants:    generateParticipants(participantCount),
			},
		})
	}

	// Generate 1-3 webinars
	webinarCount := 1 + rand.Intn(2)

	for i := 0; i < webinarCount; i++ {
		timestamp := randomTimeInRange(tr)
		durationMinutes := 30 + rand.Intn(60) // 30-90 minutes
		attendeeCount := 10 + rand.Intn(90)   // 10-100 attendees

		payload := map[string]interface{}{
			"id":       fmt.Sprintf("zoom-webinar-%d", i),
			"topic":    fmt.Sprintf("Webinar: %s", randomWebinarTopic()),
			"duration": durationMinutes,
			"host": map[string]string{
				"id": string(userID),
			},
			"attendees": attendeeCount,
		}

		payloadJSON, _ := json.Marshal(payload)

		durationSeconds := durationMinutes * 60

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("zoom-webinar-%d", i)),
			UserID:    userID,
			Source:    types.PlatformZoom,
			Type:      types.EventZoomWebinar,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:          string(userID),
				DurationSeconds: &durationSeconds,
			},
		})
	}

	// Generate 2-5 recordings
	recordingCount := 2 + rand.Intn(3)

	for i := 0; i < recordingCount; i++ {
		timestamp := randomTimeInRange(tr)

		payload := map[string]interface{}{
			"id":        fmt.Sprintf("zoom-recording-%d", i),
			"topic":     randomMeetingTopic(),
			"file_size": 1024 * 1024 * (50 + rand.Intn(200)), // 50-250 MB
			"file_type": "MP4",
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("zoom-recording-%d", i)),
			UserID:    userID,
			Source:    types.PlatformZoom,
			Type:      types.EventZoomRecording,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author: string(userID),
			},
		})
	}

	return mockEvents
}

// ============================================================================
// HELPERS
// ============================================================================

func randomMeetingTopic() string {
	topics := []string{
		"Daily Standup",
		"Sprint Planning",
		"1:1 Sync",
		"Team Retrospective",
		"Product Review",
		"Architecture Discussion",
		"Client Demo",
		"Technical Interview",
	}
	return topics[rand.Intn(len(topics))]
}

func randomMeetingType() string {
	types := []string{"scheduled", "instant", "recurring"}
	return types[rand.Intn(len(types))]
}

func randomWebinarTopic() string {
	topics := []string{
		"Product Launch Event",
		"Technical Deep Dive",
		"Industry Trends 2025",
		"Customer Success Stories",
		"Best Practices Workshop",
	}
	return topics[rand.Intn(len(topics))]
}

func generateParticipants(count int) []string {
	participants := []string{}
	for i := 0; i < count; i++ {
		participants = append(participants, fmt.Sprintf("user-%d@example.com", i))
	}
	return participants
}

func randomTimeInRange(tr types.TimeRange) time.Time {
	duration := tr.End.Sub(tr.Start)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	return tr.Start.Add(randomDuration)
}
