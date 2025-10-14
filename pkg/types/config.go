package types

import "time"

// ============================================================================
// CONFIGURATION
// ============================================================================

// AppConfig holds global application settings
type AppConfig struct {
	DatabaseURL     string
	ServerPort      int
	Environment     string // "development" | "staging" | "production"
	LogLevel        string
	EnableMetrics   bool
	EnableProfiling bool
}

// PlatformConfig holds credentials for external APIs
type PlatformConfig struct {
	Slack  SlackConfig
	GitHub GitHubConfig
	Jira   JiraConfig
	Zoom   ZoomConfig
}

// SlackConfig for Slack API
type SlackConfig struct {
	Token         string
	AppToken      string // For socket mode
	SigningSecret string
	WorkspaceID   string
	RateLimit     RateLimitConfig
	RetryConfig   RetryConfig
}

// GitHubConfig for GitHub API
type GitHubConfig struct {
	Token           string
	Organization    string
	RateLimit       RateLimitConfig
	RetryConfig     RetryConfig
	GraphQLEndpoint string
	RESTEndpoint    string
}

// JiraConfig for Jira API
type JiraConfig struct {
	URL         string // e.g., "https://company.atlassian.net"
	Email       string
	APIToken    string
	CloudID     string
	RateLimit   RateLimitConfig
	RetryConfig RetryConfig
}

// ZoomConfig for Zoom API
type ZoomConfig struct {
	AccountID    string
	ClientID     string
	ClientSecret string
	RateLimit    RateLimitConfig
	RetryConfig  RetryConfig
}

// RateLimitConfig for API rate limiting
type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
	MaxRetries        int
}

// RetryConfig for failed API calls
type RetryConfig struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// CacheConfig for connector caching
type CacheConfig struct {
	Enabled bool
	TTL     time.Duration
	MaxSize int
}
