package connectors

import (
	"github.com/vinodhalaharvi/purepulse/internal/connectors/github"
	"github.com/vinodhalaharvi/purepulse/internal/connectors/jira"
	"github.com/vinodhalaharvi/purepulse/internal/connectors/slack"
	"github.com/vinodhalaharvi/purepulse/internal/connectors/zoom"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// CONNECTOR FACTORY
// ============================================================================

// ConnectorSet holds all platform connectors
type ConnectorSet struct {
	Slack  Connector
	GitHub Connector
	Jira   Connector
	Zoom   Connector
}

// AsMap returns connectors as a map (useful for parallel fetch)
func (cs ConnectorSet) AsMap() map[types.Platform]Connector {
	return map[types.Platform]Connector{
		types.PlatformSlack:  cs.Slack,
		types.PlatformGitHub: cs.GitHub,
		types.PlatformJira:   cs.Jira,
		types.PlatformZoom:   cs.Zoom,
	}
}

// NewMockConnectors creates a full set of mock connectors
func NewMockConnectors(config types.PlatformConfig) ConnectorSet {
	// Create mock clients
	slackClient := slack.NewMockClient(config.Slack)
	githubClient := github.NewMockClient(config.GitHub)
	jiraClient := jira.NewMockClient(config.Jira)
	zoomClient := zoom.NewMockClient(config.Zoom)

	// Wrap them in Connector structs directly (no ToConnector() needed)
	return ConnectorSet{
		Slack: Connector{
			Platform: types.PlatformSlack,
			Fetch:    slackClient.Fetch,
		},
		GitHub: Connector{
			Platform: types.PlatformGitHub,
			Fetch:    githubClient.Fetch,
		},
		Jira: Connector{
			Platform: types.PlatformJira,
			Fetch:    jiraClient.Fetch,
		},
		Zoom: Connector{
			Platform: types.PlatformZoom,
			Fetch:    zoomClient.Fetch,
		},
	}
}

// NewMockConnectorsWithLatency creates mocks with custom latency
func NewMockConnectorsWithLatency(config types.PlatformConfig, latencyMS int) ConnectorSet {
	slackClient := slack.NewMockClient(config.Slack).WithLatency(latencyMS)
	githubClient := github.NewMockClient(config.GitHub).WithLatency(latencyMS)
	jiraClient := jira.NewMockClient(config.Jira).WithLatency(latencyMS)
	zoomClient := zoom.NewMockClient(config.Zoom).WithLatency(latencyMS)

	return ConnectorSet{
		Slack: Connector{
			Platform: types.PlatformSlack,
			Fetch:    slackClient.Fetch,
		},
		GitHub: Connector{
			Platform: types.PlatformGitHub,
			Fetch:    githubClient.Fetch,
		},
		Jira: Connector{
			Platform: types.PlatformJira,
			Fetch:    jiraClient.Fetch,
		},
		Zoom: Connector{
			Platform: types.PlatformZoom,
			Fetch:    zoomClient.Fetch,
		},
	}
}

// NewMockConnectorsWithFailures creates mocks with random failures
func NewMockConnectorsWithFailures(config types.PlatformConfig, failureRate float64) ConnectorSet {
	slackClient := slack.NewMockClient(config.Slack).WithFailureRate(failureRate)
	githubClient := github.NewMockClient(config.GitHub).WithFailureRate(failureRate)
	jiraClient := jira.NewMockClient(config.Jira).WithFailureRate(failureRate)
	zoomClient := zoom.NewMockClient(config.Zoom).WithFailureRate(failureRate)

	return ConnectorSet{
		Slack: Connector{
			Platform: types.PlatformSlack,
			Fetch:    slackClient.Fetch,
		},
		GitHub: Connector{
			Platform: types.PlatformGitHub,
			Fetch:    githubClient.Fetch,
		},
		Jira: Connector{
			Platform: types.PlatformJira,
			Fetch:    jiraClient.Fetch,
		},
		Zoom: Connector{
			Platform: types.PlatformZoom,
			Fetch:    zoomClient.Fetch,
		},
	}
}
