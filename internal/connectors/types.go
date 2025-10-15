package connectors

import (
	"context"

	"github.com/vinodhalaharvi/purepulse/pkg/events"
	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// CONNECTOR
// ============================================================================

// Connector interface for platform fetchers
type Connector struct {
	Platform types.Platform
	Fetch    func(ctx context.Context, userID types.UserID, tr types.TimeRange) ([]events.Event, error)
}

// ConnectorConfig for connector-specific settings
type ConnectorConfig struct {
	Platform    types.Platform
	Enabled     bool
	RateLimit   types.RateLimitConfig
	RetryConfig types.RetryConfig
	CacheConfig types.CacheConfig
}
