package events

// ============================================================================
// EVENT METADATA
// ============================================================================

// EventMetadata contains derived metadata from events
type EventMetadata struct {
	// Common across all platforms
	Author       string   `json:"author,omitempty"`
	Channel      string   `json:"channel,omitempty"` // Slack channel, GitHub repo, Jira project
	ThreadID     string   `json:"thread_id,omitempty"`
	ParentID     string   `json:"parent_id,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Participants []string `json:"participants,omitempty"`

	// Metrics
	Size            int  `json:"size,omitempty"`             // Lines of code, message length, etc.
	DurationSeconds *int `json:"duration_seconds,omitempty"` // Meeting duration, PR open time

	// Relationships
	References []string `json:"references,omitempty"` // Related events, linked issues

	// Platform-specific (kept for correlation)
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// HasAuthor checks if metadata has author
func (em EventMetadata) HasAuthor() bool {
	return em.Author != ""
}

// HasChannel checks if metadata has channel
func (em EventMetadata) HasChannel() bool {
	return em.Channel != ""
}

// IsThreaded checks if event is part of a thread
func (em EventMetadata) IsThreaded() bool {
	return em.ThreadID != "" || em.ParentID != ""
}

// ParticipantCount returns number of participants
func (em EventMetadata) ParticipantCount() int {
	return len(em.Participants)
}

// HasDuration checks if event has duration
func (em EventMetadata) HasDuration() bool {
	return em.DurationSeconds != nil && *em.DurationSeconds > 0
}
