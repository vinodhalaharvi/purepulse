package database

import (
	"time"
)

// ============================================================================
// DATABASE OPERATIONS
// ============================================================================

// InsertResult for bulk insert operations
type InsertResult struct {
	RowsInserted int           `json:"rows_inserted"`
	RowsFailed   int           `json:"rows_failed"`
	Duration     time.Duration `json:"duration"`
	Errors       []DBError     `json:"errors,omitempty"`
}

// DBError for database operation failures
type DBError struct {
	Operation string    `json:"operation"` // "insert", "select", "update"
	Table     string    `json:"table"`
	Query     string    `json:"query"`
	Err       error     `json:"error"` // ← RENAMED from Error to Err
	Timestamp time.Time `json:"timestamp"`
	Retryable bool      `json:"retryable"`
}

// Error implements error interface
func (dbe DBError) Error() string {
	if dbe.Err != nil { // ← Use Err field
		return dbe.Err.Error()
	}
	return "database error"
}
