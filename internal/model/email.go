package model

import (
	"time"

	"github.com/google/uuid"
)

type Email struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	TeamID         uuid.UUID  `json:"team_id" db:"team_id"`
	DomainID       *uuid.UUID `json:"domain_id,omitempty" db:"domain_id"`
	FromAddress    string     `json:"from" db:"from_address"`
	ToAddresses    []string   `json:"to" db:"to_addresses"`
	CcAddresses    []string   `json:"cc,omitempty" db:"cc_addresses"`
	BccAddresses   []string   `json:"bcc,omitempty" db:"bcc_addresses"`
	ReplyTo        *string    `json:"reply_to,omitempty" db:"reply_to"`
	Subject        string     `json:"subject" db:"subject"`
	HTMLBody       *string    `json:"html,omitempty" db:"html_body"`
	TextBody       *string    `json:"text,omitempty" db:"text_body"`
	Status         string     `json:"status" db:"status"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	SentAt         *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	Tags           []string   `json:"tags,omitempty" db:"tags"`
	Headers        JSONMap    `json:"headers,omitempty" db:"headers"`
	Attachments    JSONArray  `json:"attachments,omitempty" db:"attachments"`
	IdempotencyKey *string    `json:"idempotency_key,omitempty" db:"idempotency_key"`
	MessageID      *string    `json:"message_id,omitempty" db:"message_id"`
	LastError      *string    `json:"last_error,omitempty" db:"last_error"`
	RetryCount     int        `json:"retry_count" db:"retry_count"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Email status constants
const (
	EmailStatusQueued    = "queued"
	EmailStatusScheduled = "scheduled"
	EmailStatusSending   = "sending"
	EmailStatusSent      = "sent"
	EmailStatusDelivered = "delivered"
	EmailStatusBounced   = "bounced"
	EmailStatusFailed    = "failed"
	EmailStatusCancelled = "cancelled"
)

// EmailEvent tracks delivery lifecycle.
type EmailEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	EmailID   uuid.UUID `json:"email_id" db:"email_id"`
	Type      string    `json:"type" db:"type"`
	Payload   JSONMap   `json:"payload" db:"payload"`
	Recipient *string   `json:"recipient,omitempty" db:"recipient"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

const (
	EventQueued       = "queued"
	EventSent         = "sent"
	EventDelivered    = "delivered"
	EventBounced      = "bounced"
	EventOpened       = "opened"
	EventClicked      = "clicked"
	EventComplained   = "complained"
	EventUnsubscribed = "unsubscribed"
	EventFailed       = "failed"
)
