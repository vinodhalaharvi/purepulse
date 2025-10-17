package analytics

import (
	"sort"
	"time"

	"github.com/vinodhalaharvi/purepulse/db/query"
	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// CORRELATION
// ============================================================================

// Correlation represents detected cross-platform pattern
type Correlation struct {
	ID          types.CorrelationID   `json:"id"`
	UserID      types.UserID          `json:"user_id"`
	Type        types.CorrelationType `json:"type"`
	Events      []events.Event        `json:"events"`     // Related events (source + target)
	Confidence  float64               `json:"confidence"` // 0.0 - 1.0
	Description string                `json:"description"`
	TimeDelta   time.Duration         `json:"time_delta"` // Time between correlated events
	Frequency   int                   `json:"frequency"`  // How many times pattern occurred
	DetectedAt  time.Time             `json:"detected_at"`
}

// IsHighConfidence checks if correlation has high confidence (>= 0.7)
func (c Correlation) IsHighConfidence() bool {
	return c.Confidence >= 0.7
}

// IsMediumConfidence checks if correlation has medium confidence (0.5 - 0.7)
func (c Correlation) IsMediumConfidence() bool {
	return c.Confidence >= 0.5 && c.Confidence < 0.7
}

// IsLowConfidence checks if correlation has low confidence (< 0.5)
func (c Correlation) IsLowConfidence() bool {
	return c.Confidence < 0.5
}

// TimeDeltaMinutes returns time delta in minutes
func (c Correlation) TimeDeltaMinutes() int {
	return int(c.TimeDelta.Minutes())
}

// TimeDeltaHours returns time delta in hours
func (c Correlation) TimeDeltaHours() float64 {
	return c.TimeDelta.Hours()
}

// ============================================================================
// CORRELATION WINDOW
// ============================================================================

// CorrelationWindow for time-window correlation
type CorrelationWindow struct {
	Duration time.Duration `json:"duration"` // e.g., 1 hour
	Offset   time.Duration `json:"offset"`   // Slide window by this amount
}

// DefaultCorrelationWindow returns default window (1 hour duration, 15 min offset)
func DefaultCorrelationWindow() CorrelationWindow {
	return CorrelationWindow{
		Duration: 1 * time.Hour,
		Offset:   15 * time.Minute,
	}
}
func aggregateGitHubMetrics(dailyActivityRows []query.DailyActivityAgg) GitHubMetrics {
	github := GitHubMetrics{}

	for _, row := range dailyActivityRows {
		if row.Source == types.PlatformGitHub {
			github.Commits += CommitCount(row.EventCount / 2) // Changed: TotalEvents → EventCount
		}
	}

	return github
}

func aggregateJiraMetrics(dailyActivityRows []query.DailyActivityAgg) JiraMetrics {
	jira := JiraMetrics{}

	for _, row := range dailyActivityRows {
		if row.Source == types.PlatformJira {
			jira.TicketsCompleted += TicketCount(row.EventCount) // Changed: TotalEvents → EventCount
		}
	}

	return jira
}

func aggregateSlackMetrics(dailyActivityRows []query.DailyActivityAgg) SlackMetrics {
	slack := SlackMetrics{ChannelsActive: []ChannelName{}}
	totalMessages := MessageCount(0)
	channelMap := make(map[ChannelName]bool)

	for _, row := range dailyActivityRows {
		if row.Source == types.PlatformSlack {
			totalMessages += MessageCount(row.EventCount) // Changed: TotalEvents → EventCount
			for _, ch := range row.Channels {             // Changed: UniqueChannels → Channels
				channelMap[ChannelName(ch)] = true
			}
		}
	}

	slack.Messages = totalMessages
	for ch := range channelMap {
		slack.ChannelsActive = append(slack.ChannelsActive, ch)
	}
	sort.Slice(slack.ChannelsActive, func(i, j int) bool {
		return slack.ChannelsActive[i] < slack.ChannelsActive[j]
	})

	return slack
}

func aggregateMeetingMetrics(dailyActivityRows []query.DailyActivityAgg) MeetingMetrics {
	meetings := MeetingMetrics{}
	totalSeconds := 0

	for _, row := range dailyActivityRows {
		totalSeconds += row.TotalDurationSeconds // Changed: TotalDurationSec → TotalDurationSeconds
	}

	meetings.TotalMeetingHours = MeetingHours(float64(totalSeconds) / 3600.0)
	meetings.ContextSwitches = ContextSwitchCount(DetectContextSwitches(dailyActivityRows))

	return meetings
}

func DetectPeakHours(dailyActivityRows []query.DailyActivityAgg) []ActivityHour {
	hourCounts := make(map[ActivityHour]int)

	for _, row := range dailyActivityRows {
		hour := ActivityHour(row.FirstEventAt.Hour())
		hourCounts[hour] += row.EventCount // Changed: row.TotalEvents → row.EventCount
	}

	type hourCount struct {
		hour  ActivityHour
		count int
	}

	var hourCounts2 []hourCount
	for h, c := range hourCounts {
		hourCounts2 = append(hourCounts2, hourCount{h, c})
	}

	sort.Slice(hourCounts2, func(i, j int) bool {
		return hourCounts2[i].count > hourCounts2[j].count
	})

	result := make([]ActivityHour, 0)
	for i, hc := range hourCounts2 {
		if i >= 3 {
			break
		}
		result = append(result, hc.hour)
	}

	return result
}
