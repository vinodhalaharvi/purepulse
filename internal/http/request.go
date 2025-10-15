package http

import "time"

// ============================================================================
// HTTP REQUEST
// ============================================================================

// HTTPRequest wraps http.Request for applicative processing
type HTTPRequest struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	RouteParams map[string]string `json:"route_params"`
	QueryParams map[string]string `json:"query_params"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
}

// RouteParams extracted from URL path
type RouteParams struct {
	UserID string `json:"user_id,omitempty"`
	TeamID string `json:"team_id,omitempty"`
}

// QueryParams extracted from query string
type QueryParams struct {
	From     *time.Time `json:"from,omitempty"`
	To       *time.Time `json:"to,omitempty"`
	Audience string     `json:"audience,omitempty"` // For LLM: "executive" | "technical" | "team"
}
