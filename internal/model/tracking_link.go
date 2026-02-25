package model

import (
	"time"

	"github.com/google/uuid"
)

type TrackingLink struct {
	ID          uuid.UUID `json:"id" db:"id"`
	EmailID     uuid.UUID `json:"email_id" db:"email_id"`
	TeamID      uuid.UUID `json:"team_id" db:"team_id"`
	Type        string    `json:"type" db:"type"`
	OriginalURL *string   `json:"original_url,omitempty" db:"original_url"`
	Recipient   string    `json:"recipient" db:"recipient"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

const (
	TrackingTypeOpen        = "open"
	TrackingTypeClick       = "click"
	TrackingTypeUnsubscribe = "unsubscribe"
)
