package internal

import (
	"github.com/vinodhalaharvi/purepulse/internal/connectors"
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
	Slack  connectors.Connector
	GitHub connectors.Connector
	Jira   connectors.Connector
	Zoom   connectors.Connector
}

// AsMap returns connectors as a map (useful for parallel fetch)
func (cs ConnectorSet) AsMap() map[types.Platform]connectors.Connector {
	return map[types.Platform]connectors.Connector{
		types.PlatformSlack:  cs.Slack,
		types.PlatformGitHub: cs.GitHub,
		types.PlatformJira:   cs.Jira,
		types.PlatformZoom:   cs.Zoom,
	}
}

// NewMockConnectors creates a full set of mock connectors
func NewMockConnectors(config types.PlatformConfig) ConnectorSet {
	return ConnectorSet{
		Slack:  slack.NewMockClient(config.Slack).ToConnector(),
		GitHub: github.NewMockClient(config.GitHub).ToConnector(),
		Jira:   jira.NewMockClient(config.Jira).ToConnector(),
		Zoom:   zoom.NewMockClient(config.Zoom).ToConnector(),
	}
}

// NewMockConnectorsWithLatency creates mocks with custom latency
func NewMockConnectorsWithLatency(config types.PlatformConfig, latencyMS int) ConnectorSet {
	return ConnectorSet{
		Slack:  slack.NewMockClient(config.Slack).WithLatency(latencyMS).ToConnector(),
		GitHub: github.NewMockClient(config.GitHub).WithLatency(latencyMS).ToConnector(),
		Jira:   jira.NewMockClient(config.Jira).WithLatency(latencyMS).ToConnector(),
		Zoom:   zoom.NewMockClient(config.Zoom).WithLatency(latencyMS).ToConnector(),
	}
}

// NewMockConnectorsWithFailures creates mocks with random failures
func NewMockConnectorsWithFailures(config types.PlatformConfig, failureRate float64) ConnectorSet {
	return ConnectorSet{
		Slack:  slack.NewMockClient(config.Slack).WithFailureRate(failureRate).ToConnector(),
		GitHub: github.NewMockClient(config.GitHub).WithFailureRate(failureRate).ToConnector(),
		Jira:   jira.NewMockClient(config.Jira).WithFailureRate(failureRate).ToConnector(),
		Zoom:   zoom.NewMockClient(config.Zoom).WithFailureRate(failureRate).ToConnector(),
	}
}
