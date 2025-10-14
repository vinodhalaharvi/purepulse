package db

import (
	"context"
	"fmt"
	"time"
)

// HealthStatus represents database health status
type HealthStatus struct {
	Status    string        `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
	Stats     HealthStats   `json:"stats"`
	Error     string        `json:"error,omitempty"`
}

// HealthStats contains connection pool statistics
type HealthStats struct {
	MaxOpenConns      int           `json:"max_open_conns"`
	OpenConns         int           `json:"open_conns"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
}

// Health checks database health
func (c *Connection) Health(ctx context.Context) HealthStatus {
	start := time.Now()

	status := HealthStatus{
		Timestamp: start,
		Status:    "healthy",
	}

	// Ping database
	if err := c.DB.PingContext(ctx); err != nil {
		status.Status = "unhealthy"
		status.Error = fmt.Sprintf("ping failed: %v", err)
		status.Latency = time.Since(start)
		return status
	}

	status.Latency = time.Since(start)

	// Get connection pool stats
	stats := c.DB.Stats()
	status.Stats = HealthStats{
		MaxOpenConns:      stats.MaxOpenConnections,
		OpenConns:         stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}

	// Check if too many connections are waiting
	if stats.WaitCount > 100 {
		status.Status = "degraded"
		status.Error = fmt.Sprintf("high wait count: %d", stats.WaitCount)
	}

	return status
}
