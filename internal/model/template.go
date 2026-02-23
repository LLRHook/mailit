package model

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TeamID      uuid.UUID `json:"team_id" db:"team_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type TemplateVersion struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TemplateID uuid.UUID `json:"template_id" db:"template_id"`
	Version    int       `json:"version" db:"version"`
	Subject    *string   `json:"subject,omitempty" db:"subject"`
	HTMLBody   *string   `json:"html_body,omitempty" db:"html_body"`
	TextBody   *string   `json:"text_body,omitempty" db:"text_body"`
	Variables  JSONArray `json:"variables" db:"variables"`
	Published  bool      `json:"published" db:"published"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
