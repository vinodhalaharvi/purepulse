package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vinodhalaharvi/purekernels/pkg/effect"
	"github.com/vinodhalaharvi/purekernels/pkg/result"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// BuildTeamWeeklyAnalysisPrompt builds prompt from team metrics and individual reports
func BuildTeamWeeklyAnalysisPrompt(
	teamID types.TeamID,
	week types.TimeRange,
	metrics analytics.WeeklyTeamAggregationMetrics,
	userReports map[types.UserID]analytics.UserWeeklyReport,
) string {

	reportLines := []string{
		fmt.Sprintf("TEAM: %s", teamID),
		fmt.Sprintf("WEEK: %s to %s", week.Start.Format("2006-01-02"), week.End.Format("2006-01-02")),
		fmt.Sprintf("TEAM SIZE: %d members", metrics.TotalUserCount),
		fmt.Sprintf("TOTAL EVENTS: %d", metrics.TotalEventCount),
		fmt.Sprintf("WINS REPORTED: %d", metrics.TotalWinsReported),
		fmt.Sprintf("BLOCKERS REPORTED: %d", metrics.TotalBlockersReported),
		"",
		"INDIVIDUAL REPORTS SUMMARY:",
	}

	for userID, report := range userReports {
		reportLines = append(reportLines, fmt.Sprintf("  %s: %d wins, %d blocked, %d in-progress",
			userID, len(report.Wins), len(report.Blocked), len(report.InProgress)))
	}

	teamSummary := strings.Join(reportLines, "\n")

	prompt := fmt.Sprintf(`Analyze this team's weekly performance and generate a structured team report.

TEAM DATA:
%s

Based on this team activity, generate a JSON report with:
{
  "executiveSummary": "Overall team performance and trends",
  "teamVelocity": "Velocity assessment and trend",
  "blockers": [
    {
      "title": "Blocker affecting team",
      "affectedMembers": ["user1", "user2"],
      "impact": "How it affects the team",
      "priority": "high|medium|low"
    }
  ],
  "collaborationInsights": "How the team is working together",
  "topPerformers": ["user1", "user2"],
  "areasForImprovement": ["area1", "area2"],
  "teamRecommendations": [
    "Recommendation 1",
    "Recommendation 2"
  ]
}

Return ONLY valid JSON, no markdown.
`, teamSummary)

	return prompt
}

type rawTeamAnalysis struct {
	ExecutiveSummary      string           `json:"executiveSummary"`
	TeamVelocity          string           `json:"teamVelocity"`
	Blockers              []rawTeamBlocker `json:"blockers"`
	CollaborationInsights string           `json:"collaborationInsights"`
	TopPerformers         []string         `json:"topPerformers"`
	AreasForImprovement   []string         `json:"areasForImprovement"`
	TeamRecommendations   []string         `json:"teamRecommendations"`
}

type rawTeamBlocker struct {
	Title           string   `json:"title"`
	AffectedMembers []string `json:"affectedMembers"`
	Impact          string   `json:"impact"`
	Priority        string   `json:"priority"`
}

// AnalyzeTeamWeekly calls Claude to analyze team performance
func AnalyzeTeamWeekly(
	ctx context.Context,
	client *ClaudeClient,
	teamID types.TeamID,
	week types.TimeRange,
	metrics analytics.WeeklyTeamAggregationMetrics,
	userReports map[types.UserID]analytics.UserWeeklyReport,
) effect.Writer[[]string, result.Result[analytics.TeamWeeklyReport]] {

	logs := []string{fmt.Sprintf("analyze_team_weekly_started: team=%s, users=%d", teamID, metrics.TotalUserCount)}

	if len(userReports) == 0 {
		logs = append(logs, "analyze_failed: no user reports")
		return effect.NewWriter(
			result.Err[analytics.TeamWeeklyReport](fmt.Errorf("no user reports for team %s", teamID)),
			logs,
		)
	}

	prompt := BuildTeamWeeklyAnalysisPrompt(teamID, week, metrics, userReports)
	logs = append(logs, fmt.Sprintf("prompt_built: %d chars", len(prompt)))

	// Build request
	payload := struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		Messages  []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}{
		Model:     client.Model,
		MaxTokens: 2048,
	}
	payload.Messages = []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}(make([]struct {
		Role    string
		Content string
	}, 1))
	payload.Messages[0].Role = "user"
	payload.Messages[0].Content = prompt

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", client.BaseURL+"/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", client.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	startTime := time.Now()
	httpResp, err := http.DefaultClient.Do(req)
	latencyMS := int(time.Since(startTime).Milliseconds())

	if err != nil {
		logs = append(logs, fmt.Sprintf("claude_call_failed: %v", err))
		return effect.NewWriter(result.Err[analytics.TeamWeeklyReport](err), logs)
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode != 200 {
		logs = append(logs, fmt.Sprintf("claude_error: status=%d", httpResp.StatusCode))
		return effect.NewWriter(
			result.Err[analytics.TeamWeeklyReport](fmt.Errorf("Claude API error: %d", httpResp.StatusCode)),
			logs,
		)
	}

	var claudeResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	json.Unmarshal(respBody, &claudeResp)

	if len(claudeResp.Content) == 0 {
		logs = append(logs, "claude_error: no content")
		return effect.NewWriter(
			result.Err[analytics.TeamWeeklyReport](fmt.Errorf("empty Claude response")),
			logs,
		)
	}

	logs = append(logs, fmt.Sprintf("claude_response_received: tokens=%d, latency=%dms", claudeResp.Usage.OutputTokens, latencyMS))

	// Parse response
	var rawAnalysis rawTeamAnalysis
	responseText := claudeResp.Content[0].Text

	jsonStr := responseText
	if idx := strings.Index(responseText, "{"); idx >= 0 {
		if endIdx := strings.LastIndex(responseText, "}"); endIdx >= 0 {
			jsonStr = responseText[idx : endIdx+1]
		}
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawAnalysis); err != nil {
		logs = append(logs, fmt.Sprintf("parse_failed: %v", err))
		return effect.NewWriter(result.Err[analytics.TeamWeeklyReport](err), logs)
	}

	logs = append(logs, fmt.Sprintf("parse_succeeded: blockers=%d, performers=%d",
		len(rawAnalysis.Blockers), len(rawAnalysis.TopPerformers)))

	// Convert to typed report
	report := analytics.TeamWeeklyReport{
		Week:               week,
		ExecutiveSummary:   rawAnalysis.ExecutiveSummary,
		Users:              userReports,
		VelocityAnalysis:   rawAnalysis.TeamVelocity,
		CollaborationNotes: rawAnalysis.CollaborationInsights,
		Recommendations:    rawAnalysis.TeamRecommendations,
		GeneratedAt:        time.Now(),
	}

	report.TeamBlockers = make([]analytics.TeamBlockerItem, len(rawAnalysis.Blockers))
	for i, b := range rawAnalysis.Blockers {
		affectedUsers := make([]types.UserID, len(b.AffectedMembers))
		for j, member := range b.AffectedMembers {
			affectedUsers[j] = types.UserID(member)
		}
		report.TeamBlockers[i] = analytics.TeamBlockerItem{
			Title:         b.Title,
			AffectedUsers: affectedUsers,
			RiskLevel:     analytics.RiskLevel(b.Priority),
			Action:        b.Impact,
		}
	}

	logs = append(logs, "analyze_team_weekly_succeeded")
	return effect.NewWriter(result.Ok(report), logs)
}
