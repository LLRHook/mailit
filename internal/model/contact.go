package model

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	ID           uuid.UUID `json:"id" db:"id"`
	AudienceID   uuid.UUID `json:"audience_id" db:"audience_id"`
	Email        string    `json:"email" db:"email"`
	FirstName    *string   `json:"first_name,omitempty" db:"first_name"`
	LastName     *string   `json:"last_name,omitempty" db:"last_name"`
	Unsubscribed bool      `json:"unsubscribed" db:"unsubscribed"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type ContactProperty struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TeamID    uuid.UUID `json:"team_id" db:"team_id"`
	Name      string    `json:"name" db:"name"`
	Label     string    `json:"label" db:"label"`
	Type      string    `json:"type" db:"type"` // string, number, boolean, date
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ContactPropertyValue struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ContactID  uuid.UUID `json:"contact_id" db:"contact_id"`
	PropertyID uuid.UUID `json:"property_id" db:"property_id"`
	Value      *string   `json:"value,omitempty" db:"value"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
