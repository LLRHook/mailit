package model

import (
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TeamID     uuid.UUID `json:"team_id" db:"team_id"`
	Level      string    `json:"level" db:"level"`
	Message    string    `json:"message" db:"message"`
	Metadata   JSONMap   `json:"metadata" db:"metadata"`
	RequestID  *string   `json:"request_id,omitempty" db:"request_id"`
	Method     *string   `json:"method,omitempty" db:"method"`
	Path       *string   `json:"path,omitempty" db:"path"`
	StatusCode *int      `json:"status_code,omitempty" db:"status_code"`
	DurationMs *int      `json:"duration_ms,omitempty" db:"duration_ms"`
	IPAddress  *string   `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
