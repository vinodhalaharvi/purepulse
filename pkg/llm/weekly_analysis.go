// pkg/llm/weekly_analysis.go
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
	"github.com/vinodhalaharvi/purepulse/db/query"
	"github.com/vinodhalaharvi/purepulse/pkg/analytics"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// BuildWeeklyAnalysisPrompt builds prompt from activity data
func BuildWeeklyAnalysisPrompt(
	userID types.UserID,
	week types.TimeRange,
	dailyActivity []query.DailyActivityAgg,
) string {
	platformSummary := make(map[string]int)
	totalEvents := 0
	channels := make(map[string]bool)

	for _, day := range dailyActivity {
		key := string(day.Source)
		platformSummary[key] += day.EventCount
		totalEvents += day.EventCount
		for _, ch := range day.Channels {
			channels[ch] = true
		}
	}

	activityLines := []string{
		fmt.Sprintf("User: %s", userID),
		fmt.Sprintf("Week: %s to %s", week.Start.Format("2006-01-02"), week.End.Format("2006-01-02")),
		fmt.Sprintf("Total Events: %d", totalEvents),
		"",
		"Activity by Platform:",
	}

	for platform, count := range platformSummary {
		activityLines = append(activityLines, fmt.Sprintf("  - %s: %d events", platform, count))
	}

	if len(channels) > 0 {
		activityLines = append(activityLines, "")
		activityLines = append(activityLines, "Active Channels/Projects:")
		for ch := range channels {
			activityLines = append(activityLines, fmt.Sprintf("  - %s", ch))
		}
	}

	activityLines = append(activityLines, "")
	activityLines = append(activityLines, "Daily Breakdown:")

	for _, day := range dailyActivity {
		activityLines = append(activityLines,
			fmt.Sprintf("  %s (%s): %d events",
				day.Date.Format("2006-01-02"),
				day.Source,
				day.EventCount))
	}

	activitySummary := strings.Join(activityLines, "\n")

	prompt := fmt.Sprintf(`Analyze this user's weekly activity and generate a structured report about their work.

ACTIVITY DATA:
%s

Based on this activity pattern, generate a JSON report with:
{
  "executiveSummary": "2-3 sentence overview of the week",
  "wins": [
    {
      "title": "Achievement title",
      "description": "What they accomplished",
      "impact": "Why it matters"
    }
  ],
  "inProgress": [
    {
      "title": "Work item",
      "percentComplete": 60,
      "blockedBy": "dependency or empty string",
      "dueDate": "YYYY-MM-DD or empty"
    }
  ],
  "blocked": [
    {
      "title": "Blocker description",
      "blockedBy": "Root cause",
      "durationHours": 24,
      "impactLevel": "low|medium|high",
      "suggestedAction": "How to resolve"
    }
  ],
  "patterns": {
    "strengths": ["Pattern 1", "Pattern 2"],
    "opportunities": ["Area 1", "Area 2"],
    "focus": "Key focus area for next week"
  },
  "recommendations": [
    "Actionable recommendation 1",
    "Actionable recommendation 2"
  ]
}

Return ONLY valid JSON, no markdown or explanation.
`, activitySummary)

	return prompt
}

// Response types
type rawWeeklyAnalysis struct {
	ExecutiveSummary string            `json:"executiveSummary"`
	Wins             []rawWinItem      `json:"wins"`
	InProgress       []rawProgressItem `json:"inProgress"`
	Blocked          []rawBlockedItem  `json:"blocked"`
	Patterns         rawPatterns       `json:"patterns"`
	Recommendations  []string          `json:"recommendations"`
}

type rawWinItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

type rawProgressItem struct {
	Title           string `json:"title"`
	PercentComplete int    `json:"percentComplete"`
	BlockedBy       string `json:"blockedBy"`
	DueDate         string `json:"dueDate"`
}

type rawBlockedItem struct {
	Title           string `json:"title"`
	BlockedBy       string `json:"blockedBy"`
	DurationHours   int    `json:"durationHours"`
	ImpactLevel     string `json:"impactLevel"`
	SuggestedAction string `json:"suggestedAction"`
}

type rawPatterns struct {
	Strengths     []string `json:"strengths"`
	Opportunities []string `json:"opportunities"`
	Focus         string   `json:"focus"`
}

type claudeRequestPayload struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Messages  []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type claudeResponsePayload struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// AnalyzeWeeklyActivity calls Claude to analyze activity
func AnalyzeWeeklyActivity(
	ctx context.Context,
	client *ClaudeClient,
	userID types.UserID,
	week types.TimeRange,
	dailyActivity []query.DailyActivityAgg,
) effect.Writer[[]string, result.Result[analytics.UserWeeklyReport]] {

	logs := []string{fmt.Sprintf("analyze_weekly_started: user=%s, events=%d", userID, len(dailyActivity))}

	if len(dailyActivity) == 0 {
		logs = append(logs, "analyze_failed: no activity data")
		return effect.NewWriter(
			result.Err[analytics.UserWeeklyReport](fmt.Errorf("no activity data for user %s", userID)),
			logs,
		)
	}

	prompt := BuildWeeklyAnalysisPrompt(userID, week, dailyActivity)
	logs = append(logs, fmt.Sprintf("prompt_built: %d chars", len(prompt)))

	// Build request payload
	payload := claudeRequestPayload{
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

	// Create HTTP request
	req, _ := http.NewRequestWithContext(ctx, "POST", client.BaseURL+"/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", client.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	startTime := time.Now()
	httpResp, err := http.DefaultClient.Do(req)
	latencyMS := int(time.Since(startTime).Milliseconds())

	if err != nil {
		logs = append(logs, fmt.Sprintf("claude_call_failed: %v", err))
		return effect.NewWriter(result.Err[analytics.UserWeeklyReport](err), logs)
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode != 200 {
		logs = append(logs, fmt.Sprintf("claude_error: status=%d, body=%s", httpResp.StatusCode, string(respBody)))
		return effect.NewWriter(
			result.Err[analytics.UserWeeklyReport](fmt.Errorf("Claude API error: %d", httpResp.StatusCode)),
			logs,
		)
	}

	var claudeResp claudeResponsePayload
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		logs = append(logs, fmt.Sprintf("unmarshal_failed: %v", err))
		return effect.NewWriter(result.Err[analytics.UserWeeklyReport](err), logs)
	}

	if len(claudeResp.Content) == 0 {
		logs = append(logs, "claude_error: no content in response")
		return effect.NewWriter(
			result.Err[analytics.UserWeeklyReport](fmt.Errorf("empty Claude response")),
			logs,
		)
	}

	logs = append(logs, fmt.Sprintf("claude_response_received: tokens=%d, latency=%dms", claudeResp.Usage.OutputTokens, latencyMS))

	// Parse response
	var rawAnalysis rawWeeklyAnalysis
	responseText := claudeResp.Content[0].Text

	// Extract JSON
	jsonStr := responseText
	if idx := strings.Index(responseText, "{"); idx >= 0 {
		if endIdx := strings.LastIndex(responseText, "}"); endIdx >= 0 {
			jsonStr = responseText[idx : endIdx+1]
		}
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawAnalysis); err != nil {
		logs = append(logs, fmt.Sprintf("parse_failed: %v", err))
		return effect.NewWriter(result.Err[analytics.UserWeeklyReport](err), logs)
	}

	logs = append(logs, fmt.Sprintf("parse_succeeded: wins=%d, blocked=%d",
		len(rawAnalysis.Wins), len(rawAnalysis.Blocked)))

	// Convert to typed report
	report := analytics.UserWeeklyReport{
		UserID:     userID,
		Wins:       make([]analytics.WinItem, len(rawAnalysis.Wins)),
		InProgress: make([]analytics.ProgressItem, len(rawAnalysis.InProgress)),
		Blocked:    make([]analytics.BlockedItem, len(rawAnalysis.Blocked)),
		Notes:      rawAnalysis.Patterns.Focus,
	}

	for i, w := range rawAnalysis.Wins {
		report.Wins[i] = analytics.WinItem{
			Title:       w.Title,
			Description: w.Description,
			Impact:      w.Impact,
		}
	}

	for i, p := range rawAnalysis.InProgress {
		report.InProgress[i] = analytics.ProgressItem{
			Title:           p.Title,
			PercentComplete: analytics.PercentComplete(p.PercentComplete),
			Blocker:         p.BlockedBy,
		}
	}

	for i, b := range rawAnalysis.Blocked {
		report.Blocked[i] = analytics.BlockedItem{
			Title:           b.Title,
			BlockedBy:       b.BlockedBy,
			DurationHours:   b.DurationHours,
			ImpactLevel:     analytics.ImpactLevel(b.ImpactLevel),
			SuggestedAction: b.SuggestedAction,
		}
	}

	logs = append(logs, "analyze_weekly_succeeded")
	return effect.NewWriter(result.Ok(report), logs)
}
