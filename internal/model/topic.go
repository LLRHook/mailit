package model

import (
	"time"

	"github.com/google/uuid"
)

type Topic struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TeamID      uuid.UUID `json:"team_id" db:"team_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ContactTopic struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ContactID  uuid.UUID `json:"contact_id" db:"contact_id"`
	TopicID    uuid.UUID `json:"topic_id" db:"topic_id"`
	Subscribed bool      `json:"subscribed" db:"subscribed"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
