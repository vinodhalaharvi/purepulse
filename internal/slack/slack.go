package slack

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/effect"
	"github.com/vinodhalaharvi/purekernels/pkg/monoid"
	"github.com/vinodhalaharvi/purekernels/pkg/result"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
	"github.com/vinodhalaharvi/purepulse/pkg/llm"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// CORE TYPES
// ============================================================================

type JobID string
type MessageTimestamp string

type SummaryType string

const (
	SummaryTypeDaily   SummaryType = "daily"
	SummaryTypeWeekly  SummaryType = "weekly"
	SummaryTypeMonthly SummaryType = "monthly"
)

type SummarySchedule struct {
	Daily   bool
	Weekly  bool
	Monthly bool
	TimeUTC string // "09:00" format
}

type SlackWorkspace struct {
	WorkspaceID string
	TeamID      types.TeamID
	BotToken    string
	ChannelID   string
	Configured  bool
}

type ScheduledJob struct {
	ID        JobID
	TeamID    types.TeamID
	Workspace SlackWorkspace
	Schedule  SummarySchedule
	LastRunAt time.Time
	NextRunAt time.Time
}

// ============================================================================
// SLACK BLOCK KIT TYPES
// ============================================================================

type SlackBlock interface{}

type SectionBlock struct {
	Type string
	Text SlackText
}

type HeaderBlock struct {
	Type string
	Text SlackText
}

type DividerBlock struct {
	Type string
}

type SlackText struct {
	Type string
	Text string
}

type SlackMessage struct {
	Channel   string
	Blocks    []SlackBlock
	Text      string
	Timestamp time.Time
}

// ============================================================================
// MONOID
// ============================================================================

type SlackMessageMonoid struct{}

func (SlackMessageMonoid) Empty() SlackMessage {
	return SlackMessage{
		Blocks: []SlackBlock{},
		Text:   "",
	}
}

func (SlackMessageMonoid) Combine(a, b SlackMessage) SlackMessage {
	blockMonoid := monoid.NewListMonoid[SlackBlock]()
	return SlackMessage{
		Channel:   coalesceString(a.Channel, b.Channel),
		Blocks:    blockMonoid.Combine(a.Blocks, b.Blocks),
		Text:      a.Text + "\n---\n" + b.Text,
		Timestamp: laterTime(a.Timestamp, b.Timestamp),
	}
}

var _ monoid.Monoid[SlackMessage] = SlackMessageMonoid{}

// ============================================================================
// FUNCTION TYPES
// ============================================================================

type GenerateSummary func(
	ctx context.Context,
	userID types.UserID,
	summaryType SummaryType,
	timeRange types.TimeRange,
) effect.Writer[[]string, result.Result[llm.Summary]]

type PostToSlackFn func(
	ctx context.Context,
	workspace SlackWorkspace,
	message SlackMessage,
) effect.Writer[[]string, result.Result[MessageTimestamp]]

type FetchTeamMembersFn func(
	ctx context.Context,
	teamID types.TeamID,
) effect.Writer[[]string, result.Result[[]types.UserID]]

// ============================================================================
// ERROR TYPES
// ============================================================================

type SlackError struct {
	Code      string
	Message   string
	Retryable bool
}

// ============================================================================
// CONFIGURATION
// ============================================================================

type SlackIntegrationConfig struct {
	Enabled          bool
	WorkspaceID      string
	BotToken         string
	ChannelID        string
	SummarySchedules map[SummaryType]SummarySchedule
	TimeZone         string
	MaxRetries       int
	RetryDelay       time.Duration
}

// ============================================================================
// HELPER FUNCTIONS (Pure)
// ============================================================================

func coalesceString(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func laterTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func parseTimeUTC(timeStr string) [2]int {
	var hour, minute int
	fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	return [2]int{hour, minute}
}

func nextDailyTime(now time.Time, hour, minute int, tz *time.Location) time.Time {
	next := now.Add(24 * time.Hour)
	next = time.Date(next.Year(), next.Month(), next.Day(), hour, minute, 0, 0, tz)
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func nextWeeklyTime(now time.Time, hour, minute int, tz *time.Location) time.Time {
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	next := now.Add(time.Duration(daysUntilMonday) * 24 * time.Hour)
	next = time.Date(next.Year(), next.Month(), next.Day(), hour, minute, 0, 0, tz)
	return next
}

func nextMonthlyTime(now time.Time, hour, minute int, tz *time.Location) time.Time {
	next := time.Date(now.Year(), now.Month()+1, 1, hour, minute, 0, 0, tz)
	if next.Before(now) {
		next = time.Date(now.Year(), now.Month()+2, 1, hour, minute, 0, 0, tz)
	}
	return next
}

func nextScheduledTime(schedule SummarySchedule, tz *time.Location) time.Time {
	now := time.Now().In(tz)
	parts := parseTimeUTC(schedule.TimeUTC)
	hour, minute := parts[0], parts[1]

	switch {
	case schedule.Monthly:
		return nextMonthlyTime(now, hour, minute, tz)
	case schedule.Weekly:
		return nextWeeklyTime(now, hour, minute, tz)
	case schedule.Daily:
		return nextDailyTime(now, hour, minute, tz)
	default:
		return now.Add(24 * time.Hour)
	}
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// ============================================================================
// TIME RANGE CALCULATION
// ============================================================================

func TimeRangeForSummary(summaryType SummaryType) types.TimeRange {
	now := time.Now()

	switch summaryType {
	case SummaryTypeDaily:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return types.TimeRange{Start: start, End: end}

	case SummaryTypeWeekly:
		daysBack := int(now.Weekday())
		if daysBack == 0 {
			daysBack = 7
		}
		start := now.Add(-time.Duration(daysBack) * 24 * time.Hour)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		end := start.Add(7 * 24 * time.Hour)
		return types.TimeRange{Start: start, End: end}

	case SummaryTypeMonthly:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.Add(time.Duration(daysInMonth(now.Year(), now.Month())) * 24 * time.Hour)
		return types.TimeRange{Start: start, End: end}

	default:
		return types.TimeRange{Start: now, End: now}
	}
}

// ============================================================================
// BLOCK FORMATTING (Pure Functions)
// ============================================================================

func HighlightsToBlocks(highlights []string) []SlackBlock {
	if len(highlights) == 0 {
		return []SlackBlock{}
	}

	blocks := make([]SlackBlock, 0, len(highlights)+1)
	blocks = append(blocks, HeaderBlock{
		Type: "header",
		Text: SlackText{
			Type: "plain_text",
			Text: "✨ Highlights",
		},
	})

	for _, highlight := range highlights {
		blocks = append(blocks, SectionBlock{
			Type: "section",
			Text: SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("• %s", highlight),
			},
		})
	}

	return blocks
}

func InsightsToBlocks(insights []string) []SlackBlock {
	if len(insights) == 0 {
		return []SlackBlock{}
	}

	blocks := make([]SlackBlock, 0, len(insights)+1)
	blocks = append(blocks, HeaderBlock{
		Type: "header",
		Text: SlackText{
			Type: "plain_text",
			Text: "💡 Insights",
		},
	})

	for _, insight := range insights {
		blocks = append(blocks, SectionBlock{
			Type: "section",
			Text: SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("• %s", insight),
			},
		})
	}

	return blocks
}

func MetricsToBlocks(metrics analytics.StructuredMetrics) []SlackBlock {
	blocks := make([]SlackBlock, 0)
	blocks = append(blocks, HeaderBlock{
		Type: "header",
		Text: SlackText{
			Type: "plain_text",
			Text: "📊 Metrics",
		},
	})

	metricsText := fmt.Sprintf(
		"*Activity Overview*\n- Total Events: %d\n- Slack Messages: %d\n- GitHub Commits: %d\n- GitHub PRs: %d\n- Jira Issues: %d\n- Meetings: %d\n\n*Scores*\n- Collaboration: %.1f%%\n- Code Quality: %.1f\n- Productivity Hours: %.1f",
		metrics.TotalActivity(),
		metrics.SlackMessages,
		metrics.GitHubCommits,
		metrics.GitHubPRs,
		metrics.JiraIssuesCreated+metrics.JiraIssuesCompleted,
		metrics.ZoomMeetings,
		metrics.CollaborationScore,
		metrics.CodeQuality,
		metrics.ActiveHours,
	)

	blocks = append(blocks, SectionBlock{
		Type: "section",
		Text: SlackText{
			Type: "mrkdwn",
			Text: metricsText,
		},
	})

	return blocks
}

func FormatSummaryToBlocks(summary llm.Summary, summaryType SummaryType) []SlackBlock {
	blocks := make([]SlackBlock, 0)

	titleText := "Daily Summary"
	switch summaryType {
	case SummaryTypeWeekly:
		titleText = "Weekly Summary"
	case SummaryTypeMonthly:
		titleText = "Monthly Summary"
	}

	blocks = append(blocks, HeaderBlock{
		Type: "header",
		Text: SlackText{
			Type: "plain_text",
			Text: titleText,
		},
	})

	blocks = append(blocks, SectionBlock{
		Type: "section",
		Text: SlackText{
			Type: "mrkdwn",
			Text: summary.AISummary.Content,
		},
	})

	blocks = append(blocks, DividerBlock{Type: "divider"})
	blocks = append(blocks, HighlightsToBlocks(summary.AISummary.Highlights)...)
	blocks = append(blocks, DividerBlock{Type: "divider"})
	blocks = append(blocks, InsightsToBlocks(summary.AISummary.Insights)...)

	return blocks
}

func BuildSlackMessage(blocks []SlackBlock, summary llm.Summary, workspace SlackWorkspace) SlackMessage {
	return SlackMessage{
		Channel:   workspace.ChannelID,
		Blocks:    blocks,
		Text:      summary.AISummary.Content,
		Timestamp: time.Now(),
	}
}

// ============================================================================
// BATCH OPERATIONS
// ============================================================================

func GenerateTeamSummaries(
	ctx context.Context,
	userIDs []types.UserID,
	summaryType SummaryType,
	timeRange types.TimeRange,
	generateSummary GenerateSummary,
) effect.Writer[[]string, result.Result[[]llm.Summary]] {

	logs := []string{
		fmt.Sprintf("generate_team_summaries_started: users=%d, type=%s", len(userIDs), summaryType),
	}

	summaries := make([]llm.Summary, 0, len(userIDs))
	allLogs := logs

	for _, userID := range userIDs {
		writer := generateSummary(ctx, userID, summaryType, timeRange)
		res, genLogs := writer.Run()
		allLogs = append(allLogs, genLogs...)

		if res.IsOk() {
			summaries = append(summaries, res.Unwrap())
		} else {
			allLogs = append(allLogs, fmt.Sprintf("generate_failed: user=%s, error=%v", userID, res.Error()))
		}
	}

	allLogs = append(allLogs, fmt.Sprintf("generate_team_summaries_completed: succeeded=%d, failed=%d",
		len(summaries), len(userIDs)-len(summaries)))

	if len(summaries) == 0 {
		return effect.NewWriter(
			result.Err[[]llm.Summary](fmt.Errorf("no summaries generated")),
			allLogs,
		)
	}

	return effect.NewWriter(result.Ok(summaries), allLogs)
}

func PostAllSummaries(
	ctx context.Context,
	workspace SlackWorkspace,
	summaries []llm.Summary,
	summaryType SummaryType,
	postToSlack PostToSlackFn,
) effect.Writer[[]string, result.Result[[]MessageTimestamp]] {

	logs := []string{
		fmt.Sprintf("post_all_summaries_started: count=%d", len(summaries)),
	}

	timestamps := make([]MessageTimestamp, 0, len(summaries))
	allLogs := logs

	for i, summary := range summaries {
		blocks := FormatSummaryToBlocks(summary, summaryType)
		message := BuildSlackMessage(blocks, summary, workspace)

		writer := postToSlack(ctx, workspace, message)
		res, postLogs := writer.Run()
		allLogs = append(allLogs, postLogs...)

		if res.IsOk() {
			timestamps = append(timestamps, res.Unwrap())
		} else {
			allLogs = append(allLogs, fmt.Sprintf("post_failed: index=%d, error=%v", i, res.Error()))
		}
	}

	allLogs = append(allLogs, fmt.Sprintf("post_all_summaries_completed: succeeded=%d, failed=%d",
		len(timestamps), len(summaries)-len(timestamps)))

	if len(timestamps) == 0 {
		return effect.NewWriter(
			result.Err[[]MessageTimestamp](fmt.Errorf("no messages posted")),
			allLogs,
		)
	}

	return effect.NewWriter(result.Ok(timestamps), allLogs)
}

// ============================================================================
// COMPLETE PIPELINE
// ============================================================================

func SummaryPipeline(
	ctx context.Context,
	workspace SlackWorkspace,
	summaryType SummaryType,
	fetchMembers FetchTeamMembersFn,
	generateSummary GenerateSummary,
	postFunc PostToSlackFn,
) effect.Writer[[]string, result.Result[[]MessageTimestamp]] {

	logs := []string{
		fmt.Sprintf("pipeline_started: workspace=%s, type=%s", workspace.WorkspaceID, summaryType),
	}

	membersWriter := fetchMembers(ctx, workspace.TeamID)
	membersRes, memberLogs := membersWriter.Run()
	logs = append(logs, memberLogs...)

	if !membersRes.IsOk() {
		logs = append(logs, fmt.Sprintf("pipeline_failed: fetch_members error=%v", membersRes.Error()))
		return effect.NewWriter(
			result.Err[[]MessageTimestamp](membersRes.Error()),
			logs,
		)
	}

	members := membersRes.Unwrap()
	timeRange := TimeRangeForSummary(summaryType)

	summariesWriter := GenerateTeamSummaries(ctx, members, summaryType, timeRange, generateSummary)
	summariesRes, genLogs := summariesWriter.Run()
	logs = append(logs, genLogs...)

	if !summariesRes.IsOk() {
		logs = append(logs, fmt.Sprintf("pipeline_failed: generate_summaries error=%v", summariesRes.Error()))
		return effect.NewWriter(
			result.Err[[]MessageTimestamp](summariesRes.Error()),
			logs,
		)
	}

	summaries := summariesRes.Unwrap()

	postWriter := PostAllSummaries(ctx, workspace, summaries, summaryType, postFunc)
	postRes, postLogs := postWriter.Run()
	logs = append(logs, postLogs...)

	if !postRes.IsOk() {
		logs = append(logs, fmt.Sprintf("pipeline_failed: post_summaries error=%v", postRes.Error()))
		return effect.NewWriter(
			result.Err[[]MessageTimestamp](postRes.Error()),
			logs,
		)
	}

	timestamps := postRes.Unwrap()
	logs = append(logs, fmt.Sprintf("pipeline_completed: messages=%d", len(timestamps)))

	return effect.NewWriter(result.Ok(timestamps), logs)
}

// ============================================================================
// SCHEDULER SERVICE
// ============================================================================

type SchedulerService struct {
	jobs         map[string]ScheduledJob
	mu           sync.RWMutex
	fetchMembers FetchTeamMembersFn
}

func NewSchedulerService(fetchMembers FetchTeamMembersFn) *SchedulerService {
	return &SchedulerService{
		jobs:         make(map[string]ScheduledJob),
		fetchMembers: fetchMembers,
	}
}

func (s *SchedulerService) Register(
	ctx context.Context,
	job ScheduledJob,
) effect.Writer[[]string, result.Result[JobID]] {

	logs := []string{
		fmt.Sprintf("register_started: job_id=%s, team=%s", job.ID, job.TeamID),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[string(job.ID)]; exists {
		logs = append(logs, "register_failed: job already exists")
		return effect.NewWriter(
			result.Err[JobID](fmt.Errorf("job already exists: %s", job.ID)),
			logs,
		)
	}

	job.NextRunAt = nextScheduledTime(job.Schedule, time.UTC)
	s.jobs[string(job.ID)] = job

	logs = append(logs, fmt.Sprintf("register_succeeded: next_run=%s", job.NextRunAt))
	return effect.NewWriter(result.Ok(job.ID), logs)
}

func (s *SchedulerService) Unregister(
	ctx context.Context,
	jobID JobID,
) effect.Writer[[]string, result.Result[struct{}]] {

	logs := []string{
		fmt.Sprintf("unregister_started: job_id=%s", jobID),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[string(jobID)]; !exists {
		logs = append(logs, "unregister_failed: job not found")
		return effect.NewWriter(
			result.Err[struct{}](fmt.Errorf("job not found: %s", jobID)),
			logs,
		)
	}

	delete(s.jobs, string(jobID))
	logs = append(logs, "unregister_succeeded")
	return effect.NewWriter(result.Ok(struct{}{}), logs)
}

func (s *SchedulerService) NextJob(
	ctx context.Context,
) effect.Writer[[]string, result.Result[*ScheduledJob]] {

	logs := []string{"next_job_started"}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var nextJob *ScheduledJob
	var earliestTime time.Time

	for _, job := range s.jobs {
		if nextJob == nil || job.NextRunAt.Before(earliestTime) {
			jobCopy := job
			nextJob = &jobCopy
			earliestTime = job.NextRunAt
		}
	}

	if nextJob == nil || nextJob.NextRunAt.After(time.Now()) {
		logs = append(logs, "next_job_not_ready")
		return effect.NewWriter(
			result.Err[*ScheduledJob](fmt.Errorf("no job ready to run")),
			logs,
		)
	}

	logs = append(logs, fmt.Sprintf("next_job_found: id=%s", nextJob.ID))
	return effect.NewWriter(result.Ok(nextJob), logs)
}

func (s *SchedulerService) ExecuteJob(
	ctx context.Context,
	job ScheduledJob,
) effect.Writer[[]string, result.Result[[]MessageTimestamp]] {

	logs := []string{
		fmt.Sprintf("execute_job_started: id=%s", job.ID),
	}

	timestamps := []MessageTimestamp{}

	s.mu.Lock()
	if j, exists := s.jobs[string(job.ID)]; exists {
		j.LastRunAt = time.Now()
		j.NextRunAt = nextScheduledTime(j.Schedule, time.UTC)
		s.jobs[string(job.ID)] = j
	}
	s.mu.Unlock()

	logs = append(logs, "execute_job_succeeded")
	return effect.NewWriter(result.Ok(timestamps), logs)
}
