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
		latencyMS:   70 + rand.Intn(130),
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

	// Realistic meeting configurations
	meetings := []struct {
		topic        string
		duration     int
		recurring    bool
		minAttendees int
		maxAttendees int
	}{
		{"Daily Standup", 15, true, 3, 8},
		{"Sprint Planning", 120, false, 5, 10},
		{"Sprint Retrospective", 60, false, 5, 10},
		{"Sprint Review/Demo", 60, false, 8, 15},
		{"1:1 with Manager", 30, true, 2, 2},
		{"Architecture Review", 90, false, 4, 8},
		{"Technical Interview", 60, false, 2, 4},
		{"Team Sync", 30, true, 4, 8},
		{"Product Roadmap Review", 60, false, 6, 12},
		{"Incident Post-Mortem", 45, false, 5, 10},
		{"Code Review Session", 45, false, 2, 5},
		{"Design Review", 60, false, 4, 8},
		{"Customer Success Sync", 30, false, 3, 6},
		{"Engineering All-Hands", 60, false, 20, 50},
		{"Security Review", 45, false, 3, 6},
		{"Performance Review", 45, false, 2, 2},
		{"Release Planning", 90, false, 6, 12},
		{"Tech Talk: Kubernetes Best Practices", 60, false, 10, 30},
		{"Onboarding Session", 60, false, 2, 3},
		{"Cross-Team Collaboration", 45, false, 6, 12},
	}

	// Generate 5-15 meetings over the time range
	meetingCount := 5 + rand.Intn(10)

	for i := 0; i < meetingCount; i++ {
		timestamp := randomTimeInRange(tr)

		// Skip weekends for most meetings
		if timestamp.Weekday() == time.Saturday || timestamp.Weekday() == time.Sunday {
			if rand.Float32() > 0.1 { // 90% skip weekends
				continue
			}
		}

		meeting := meetings[rand.Intn(len(meetings))]
		participantCount := meeting.minAttendees + rand.Intn(meeting.maxAttendees-meeting.minAttendees+1)
		meetingID := fmt.Sprintf("%09d", 90000000+rand.Intn(9999999))

		payload := map[string]interface{}{
			"id":       meetingID,
			"uuid":     fmt.Sprintf("v4uuid-%s", meetingID),
			"topic":    meeting.topic,
			"duration": meeting.duration,
			"host": map[string]string{
				"id":    string(userID),
				"email": fmt.Sprintf("%s@example.com", userID),
			},
			"participants":    participantCount,
			"type":            getMeetingType(meeting.recurring),
			"start_time":      timestamp.Format(time.RFC3339),
			"timezone":        "America/New_York",
			"join_url":        fmt.Sprintf("https://zoom.us/j/%s", meetingID),
			"recording_count": 0,
		}

		payloadJSON, _ := json.Marshal(payload)

		durationSeconds := meeting.duration * 60

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("zoom-meeting-%s-%d", userID, i)),
			UserID:    userID,
			Source:    types.PlatformZoom,
			Type:      types.EventZoomMeeting,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:          string(userID),
				Channel:         meeting.topic,
				DurationSeconds: &durationSeconds,
				Participants:    generateParticipants(participantCount),
			},
		})
	}

	// Generate 0-2 webinars (less common)
	webinarCount := rand.Intn(3)
	webinarTopics := []string{
		"Q4 Product Roadmap Presentation",
		"Engineering Best Practices Workshop",
		"Customer Success Stories",
		"Platform Migration Strategy",
		"Security Compliance Training",
		"New Feature Launch Demo",
		"Technical Deep Dive: Microservices",
		"Annual Technology Trends",
	}

	for i := 0; i < webinarCount; i++ {
		timestamp := randomTimeInRange(tr)
		durationMinutes := 45 + rand.Intn(45) // 45-90 minutes
		attendeeCount := 20 + rand.Intn(80)   // 20-100 attendees
		webinarID := fmt.Sprintf("%09d", 95000000+rand.Intn(999999))

		payload := map[string]interface{}{
			"id":       webinarID,
			"topic":    webinarTopics[rand.Intn(len(webinarTopics))],
			"duration": durationMinutes,
			"host": map[string]string{
				"id":    string(userID),
				"email": fmt.Sprintf("%s@example.com", userID),
			},
			"attendees":   attendeeCount,
			"registrants": attendeeCount + rand.Intn(50),
			"type":        "webinar",
		}

		payloadJSON, _ := json.Marshal(payload)

		durationSeconds := durationMinutes * 60

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("zoom-webinar-%s-%d", userID, i)),
			UserID:    userID,
			Source:    types.PlatformZoom,
			Type:      types.EventZoomWebinar,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:          string(userID),
				Channel:         "webinar",
				DurationSeconds: &durationSeconds,
			},
		})
	}

	// Generate 2-5 recordings (from important meetings)
	recordingCount := 2 + rand.Intn(3)
	recordableTopics := []string{
		"Sprint Planning Recording",
		"Architecture Review Recording",
		"Tech Talk Recording",
		"All-Hands Recording",
		"Training Session Recording",
		"Customer Demo Recording",
		"Product Review Recording",
	}

	for i := 0; i < recordingCount; i++ {
		timestamp := randomTimeInRange(tr)
		fileSizeMB := 100 + rand.Intn(400) // 100-500 MB
		recordingID := fmt.Sprintf("rec_%d", rand.Int63n(999999999))

		payload := map[string]interface{}{
			"id":              recordingID,
			"topic":           recordableTopics[rand.Intn(len(recordableTopics))],
			"file_size":       fileSizeMB * 1024 * 1024,
			"file_type":       "MP4",
			"recording_start": timestamp.Format(time.RFC3339),
			"duration":        30 + rand.Intn(90), // 30-120 minutes
			"download_url":    fmt.Sprintf("https://zoom.us/rec/download/%s", recordingID),
			"status":          "completed",
		}

		payloadJSON, _ := json.Marshal(payload)

		mockEvents = append(mockEvents, events.Event{
			ID:        types.EventID(fmt.Sprintf("zoom-recording-%s-%d", userID, i)),
			UserID:    userID,
			Source:    types.PlatformZoom,
			Type:      types.EventZoomRecording,
			Timestamp: timestamp,
			Payload:   payloadJSON,
			Metadata: events.EventMetadata{
				Author:  string(userID),
				Channel: "recording",
				Size:    fileSizeMB * 1024 * 1024,
			},
		})
	}

	return mockEvents
}

// Helper functions
func getMeetingType(recurring bool) string {
	if recurring {
		return "recurring"
	}
	if rand.Float32() < 0.3 {
		return "instant"
	}
	return "scheduled"
}

func generateParticipants(count int) []string {
	participants := []string{}
	names := []string{"alice", "bob", "charlie", "david", "emma", "frank", "grace", "henry", "iris", "jack"}
	for i := 0; i < count && i < len(names); i++ {
		participants = append(participants, fmt.Sprintf("%s@example.com", names[i]))
	}
	return participants
}

func randomTimeInRange(tr types.TimeRange) time.Time {
	duration := tr.End.Sub(tr.Start)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	return tr.Start.Add(randomDuration)
}
