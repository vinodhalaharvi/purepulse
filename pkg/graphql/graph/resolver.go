package graph

import (
	"database/sql"

	"github.com/vinodhalaharvi/purepulse/pkg/llm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB           *sql.DB
	ClaudeClient *llm.ClaudeClient
}
