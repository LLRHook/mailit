package model

import (
	"time"

	"github.com/google/uuid"
)

type ContactImportJob struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TeamID        uuid.UUID `json:"team_id" db:"team_id"`
	AudienceID    uuid.UUID `json:"audience_id" db:"audience_id"`
	Status        string    `json:"status" db:"status"`
	TotalRows     int       `json:"total_rows" db:"total_rows"`
	ProcessedRows int       `json:"processed_rows" db:"processed_rows"`
	CreatedRows   int       `json:"created_rows" db:"created_rows"`
	UpdatedRows   int       `json:"updated_rows" db:"updated_rows"`
	SkippedRows   int       `json:"skipped_rows" db:"skipped_rows"`
	FailedRows    int       `json:"failed_rows" db:"failed_rows"`
	Error         *string   `json:"error,omitempty" db:"error"`
	CSVData       string    `json:"-" db:"csv_data"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

const (
	ImportStatusPending    = "pending"
	ImportStatusProcessing = "processing"
	ImportStatusCompleted  = "completed"
	ImportStatusFailed     = "failed"
)
