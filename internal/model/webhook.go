package model

import (
	"time"

	"github.com/google/uuid"
)

type Webhook struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TeamID        uuid.UUID `json:"team_id" db:"team_id"`
	URL           string    `json:"url" db:"url"`
	Events        []string  `json:"events" db:"events"`
	SigningSecret string    `json:"-" db:"signing_secret"`
	Active        bool      `json:"active" db:"active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type WebhookEvent struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	WebhookID    uuid.UUID  `json:"webhook_id" db:"webhook_id"`
	EventType    string     `json:"event_type" db:"event_type"`
	Payload      JSONMap    `json:"payload" db:"payload"`
	Status       string     `json:"status" db:"status"`
	ResponseCode *int       `json:"response_code,omitempty" db:"response_code"`
	ResponseBody *string    `json:"response_body,omitempty" db:"response_body"`
	Attempts     int        `json:"attempts" db:"attempts"`
	NextRetryAt  *time.Time `json:"next_retry_at,omitempty" db:"next_retry_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}
