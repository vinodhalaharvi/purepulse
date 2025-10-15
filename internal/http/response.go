package http

// ============================================================================
// HTTP RESPONSE
// ============================================================================

// HTTPResponse for applicative response building
type HTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

// IsSuccess checks if response is successful (2xx)
func (hr HTTPResponse) IsSuccess() bool {
	return hr.StatusCode >= 200 && hr.StatusCode < 300
}

// IsError checks if response is error (4xx or 5xx)
func (hr HTTPResponse) IsError() bool {
	return hr.StatusCode >= 400
}
