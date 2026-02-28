package model

import (
	"time"

	"github.com/google/uuid"
)

type Broadcast struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	TeamID          uuid.UUID  `json:"team_id" db:"team_id"`
	Name            string     `json:"name" db:"name"`
	AudienceID      *uuid.UUID `json:"audience_id,omitempty" db:"audience_id"`
	SegmentID       *uuid.UUID `json:"segment_id,omitempty" db:"segment_id"`
	TemplateID      *uuid.UUID `json:"template_id,omitempty" db:"template_id"`
	FromAddress     *string    `json:"from,omitempty" db:"from_address"`
	Subject         *string    `json:"subject,omitempty" db:"subject"`
	HTMLBody        *string    `json:"html_body,omitempty" db:"html_body"`
	TextBody        *string    `json:"text_body,omitempty" db:"text_body"`
	Status          string     `json:"status" db:"status"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	SentAt          *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	TotalRecipients int        `json:"total_recipients" db:"total_recipients"`
	SentCount       int        `json:"sent_count" db:"sent_count"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	AudienceName    *string    `json:"-" db:"-"` // populated via JOIN, not a DB column
}

const (
	BroadcastStatusDraft     = "draft"
	BroadcastStatusQueued    = "queued"
	BroadcastStatusSending   = "sending"
	BroadcastStatusSent      = "sent"
	BroadcastStatusCancelled = "cancelled"
)
