package model

import (
	"time"

	"github.com/google/uuid"
)

type Segment struct {
	ID         uuid.UUID `json:"id" db:"id"`
	AudienceID uuid.UUID `json:"audience_id" db:"audience_id"`
	Name       string    `json:"name" db:"name"`
	Conditions JSONArray `json:"conditions" db:"conditions"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type SegmentContact struct {
	SegmentID uuid.UUID `json:"segment_id" db:"segment_id"`
	ContactID uuid.UUID `json:"contact_id" db:"contact_id"`
	AddedAt   time.Time `json:"added_at" db:"added_at"`
}
