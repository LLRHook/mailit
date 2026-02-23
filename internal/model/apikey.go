package model

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TeamID     uuid.UUID  `json:"team_id" db:"team_id"`
	Name       string     `json:"name" db:"name"`
	KeyHash    string     `json:"-" db:"key_hash"`
	KeyPrefix  string     `json:"key_prefix" db:"key_prefix"`
	Permission string     `json:"permission" db:"permission"`
	DomainID   *uuid.UUID `json:"domain_id,omitempty" db:"domain_id"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

const (
	PermissionFull    = "full"
	PermissionSending = "sending"
)
