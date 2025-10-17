package analytics

import (
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// MEANINGFUL TYPE ALIASES
// ============================================================================

type CommitCount int
type QualityScore float64
type ReviewTurnaroundMS int
type LinesCount int

type TicketCount int
type StoryPoints float64
type CompletionDays float64

type MessageCount int
type ChannelName string
type PairProgrammingHours float64
type HelpQuestionsAnswered int

type MeetingHours float64
type DeepWorkHours float64
type ContextSwitchCount int
type ActivityHour int // 0-23

type CollaborationScore float64
type ProductivityScore float64
type TeamVelocity int

type PercentComplete int // 0-100
type ImpactLevel string  // "low", "medium", "high"
type BlockerDurationDays int
type RiskLevel string

// ============================================================================
// GITHUB METRICS
// ============================================================================

type GitHubMetrics struct {
	Commits            CommitCount
	CommitsAvgQuality  QualityScore
	PRsOpened          int
	PRsReviewed        int
	ReviewTurnaroundMS ReviewTurnaroundMS
	LinesAdded         LinesCount
	LinesDeleted       LinesCount
}

// ============================================================================
// JIRA METRICS
// ============================================================================

type JiraMetrics struct {
	TicketsCompleted     TicketCount
	TicketsInProgress    TicketCount
	TicketsBlocked       TicketCount
	AvgCompletionDays    CompletionDays
	StoryPointsCompleted StoryPoints
}

// ============================================================================
// SLACK METRICS
// ============================================================================

type SlackMetrics struct {
	Messages              MessageCount
	ChannelsActive        []ChannelName
	PairProgrammingHours  PairProgrammingHours
	HelpQuestionsAnswered HelpQuestionsAnswered
}

// ============================================================================
// MEETING METRICS
// ============================================================================

type MeetingMetrics struct {
	TotalMeetingHours MeetingHours
	DeepWorkHours     DeepWorkHours
	ContextSwitches   ContextSwitchCount
}

// ============================================================================
// ACTIVITY PATTERNS
// ============================================================================

type ActivityPatterns struct {
	PeakActivityHours  []ActivityHour
	CollaborationScore CollaborationScore
	ProductivityScore  ProductivityScore
}

// ============================================================================
// USER METRICS
// ============================================================================

type WeeklyUserMetrics struct {
	UserID      types.UserID
	Week        types.TimeRange
	GitHub      GitHubMetrics
	Jira        JiraMetrics
	Slack       SlackMetrics
	Meetings    MeetingMetrics
	Patterns    ActivityPatterns
	GeneratedAt time.Time
}

// ============================================================================
// TEAM METRICS
// ============================================================================

type MentorshipRelation struct {
	Mentor               types.UserID
	Mentee               types.UserID
	PairProgrammingHours PairProgrammingHours
	FocusArea            string
}

type WeeklyTeamMetrics struct {
	TeamID             types.TeamID
	Week               types.TimeRange
	Members            []WeeklyUserMetrics
	Velocity           TeamVelocity
	VelocityTrend      float64 // percentage
	BlockerCount       int
	AvgBlockerDuration CompletionDays
	CollaborationScore CollaborationScore
	TopCollaborators   []types.UserID
	KeyBlockers        []string
	MentorshipPairs    []MentorshipRelation
	GeneratedAt        time.Time
}

// ============================================================================
// REPORT ITEMS
// ============================================================================

type WinItem struct {
	Title       string
	Description string
	Impact      string
}

type ProgressItem struct {
	Title           string
	PercentComplete PercentComplete
	Blocker         string
	DueDate         time.Time
}

type BlockedItem struct {
	Title           string
	BlockedBy       string
	DurationHours   int
	ImpactLevel     ImpactLevel
	SuggestedAction string
}

type UserWeeklyReport struct {
	UserID     types.UserID
	Wins       []WinItem
	InProgress []ProgressItem
	Blocked    []BlockedItem
	Notes      string
}

type TeamBlockerItem struct {
	Title         string
	AffectedUsers []types.UserID
	DurationDays  BlockerDurationDays
	RiskLevel     RiskLevel
	Action        string
}

type MetricTrend struct {
	Current   interface{}
	Previous  interface{}
	Trend     float64 // percentage
	Direction string  // "up", "down", "stable"
}

type TeamWeeklyReport struct {
	Week                types.TimeRange
	ExecutiveSummary    string
	Users               map[types.UserID]UserWeeklyReport
	VelocityAnalysis    string
	CollaborationNotes  string
	TeamBlockers        []TeamBlockerItem
	Recommendations     []string
	GrowthOpportunities []string
	MetricsSummary      map[string]MetricTrend
	AIModel             string
	AITokens            int
	AILatencyMS         int
	GeneratedAt         time.Time
}

// ============================================================================
// MONOID INSTANCES
// ============================================================================

type WeeklyUserMetricsMonoid struct{}

var _ monoid.Monoid[WeeklyUserMetrics] = WeeklyUserMetricsMonoid{}

func (WeeklyUserMetricsMonoid) Empty() WeeklyUserMetrics {
	return WeeklyUserMetrics{
		GitHub:      GitHubMetrics{},
		Jira:        JiraMetrics{},
		Slack:       SlackMetrics{ChannelsActive: []ChannelName{}},
		Meetings:    MeetingMetrics{},
		Patterns:    ActivityPatterns{PeakActivityHours: []ActivityHour{}},
		GeneratedAt: time.Now(),
	}
}

func (WeeklyUserMetricsMonoid) Combine(a, b WeeklyUserMetrics) WeeklyUserMetrics {
	return WeeklyUserMetrics{
		UserID: a.UserID,
		Week:   a.Week,
		GitHub: GitHubMetrics{
			Commits:            a.GitHub.Commits + b.GitHub.Commits,
			CommitsAvgQuality:  (a.GitHub.CommitsAvgQuality + b.GitHub.CommitsAvgQuality) / 2,
			PRsOpened:          a.GitHub.PRsOpened + b.GitHub.PRsOpened,
			PRsReviewed:        a.GitHub.PRsReviewed + b.GitHub.PRsReviewed,
			ReviewTurnaroundMS: (a.GitHub.ReviewTurnaroundMS + b.GitHub.ReviewTurnaroundMS) / 2,
			LinesAdded:         a.GitHub.LinesAdded + b.GitHub.LinesAdded,
			LinesDeleted:       a.GitHub.LinesDeleted + b.GitHub.LinesDeleted,
		},
		Jira: JiraMetrics{
			TicketsCompleted:     a.Jira.TicketsCompleted + b.Jira.TicketsCompleted,
			TicketsInProgress:    a.Jira.TicketsInProgress + b.Jira.TicketsInProgress,
			TicketsBlocked:       a.Jira.TicketsBlocked + b.Jira.TicketsBlocked,
			AvgCompletionDays:    (a.Jira.AvgCompletionDays + b.Jira.AvgCompletionDays) / 2,
			StoryPointsCompleted: a.Jira.StoryPointsCompleted + b.Jira.StoryPointsCompleted,
		},
		Slack: SlackMetrics{
			Messages:              a.Slack.Messages + b.Slack.Messages,
			ChannelsActive:        append(a.Slack.ChannelsActive, b.Slack.ChannelsActive...),
			PairProgrammingHours:  a.Slack.PairProgrammingHours + b.Slack.PairProgrammingHours,
			HelpQuestionsAnswered: a.Slack.HelpQuestionsAnswered + b.Slack.HelpQuestionsAnswered,
		},
		Meetings: MeetingMetrics{
			TotalMeetingHours: a.Meetings.TotalMeetingHours + b.Meetings.TotalMeetingHours,
			DeepWorkHours:     a.Meetings.DeepWorkHours + b.Meetings.DeepWorkHours,
			ContextSwitches:   a.Meetings.ContextSwitches + b.Meetings.ContextSwitches,
		},
		GeneratedAt: time.Now(),
	}
}

type WeeklyTeamMetricsMonoid struct{}

var _ monoid.Monoid[WeeklyTeamMetrics] = WeeklyTeamMetricsMonoid{}

func (WeeklyTeamMetricsMonoid) Empty() WeeklyTeamMetrics {
	return WeeklyTeamMetrics{
		Members:          []WeeklyUserMetrics{},
		TopCollaborators: []types.UserID{},
		KeyBlockers:      []string{},
		MentorshipPairs:  []MentorshipRelation{},
		GeneratedAt:      time.Now(),
	}
}

func (WeeklyTeamMetricsMonoid) Combine(a, b WeeklyTeamMetrics) WeeklyTeamMetrics {
	return WeeklyTeamMetrics{
		TeamID:             a.TeamID,
		Week:               a.Week,
		Members:            append(a.Members, b.Members...),
		Velocity:           a.Velocity + b.Velocity,
		VelocityTrend:      (a.VelocityTrend + b.VelocityTrend) / 2,
		BlockerCount:       a.BlockerCount + b.BlockerCount,
		AvgBlockerDuration: (a.AvgBlockerDuration + b.AvgBlockerDuration) / 2,
		CollaborationScore: (a.CollaborationScore + b.CollaborationScore) / 2,
		GeneratedAt:        time.Now(),
	}
}
