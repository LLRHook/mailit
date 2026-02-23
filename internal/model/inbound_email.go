package model

import (
	"time"

	"github.com/google/uuid"
)

type InboundEmail struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TeamID      uuid.UUID  `json:"team_id" db:"team_id"`
	DomainID    *uuid.UUID `json:"domain_id,omitempty" db:"domain_id"`
	FromAddress string     `json:"from" db:"from_address"`
	ToAddresses []string   `json:"to" db:"to_addresses"`
	CcAddresses []string   `json:"cc,omitempty" db:"cc_addresses"`
	Subject     *string    `json:"subject,omitempty" db:"subject"`
	HTMLBody    *string    `json:"html_body,omitempty" db:"html_body"`
	TextBody    *string    `json:"text_body,omitempty" db:"text_body"`
	RawMessage  *string    `json:"raw_message,omitempty" db:"raw_message"`
	Headers     JSONMap    `json:"headers" db:"headers"`
	Attachments JSONArray  `json:"attachments" db:"attachments"`
	SpamScore   *float64   `json:"spam_score,omitempty" db:"spam_score"`
	Processed   bool       `json:"processed" db:"processed"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}
