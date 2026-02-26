package dto

type CreateBroadcastRequest struct {
	Name       string  `json:"name" validate:"required"`
	AudienceID *string `json:"audience_id,omitempty" validate:"omitempty,uuid"`
	SegmentID  *string `json:"segment_id,omitempty" validate:"omitempty,uuid"`
	TemplateID *string `json:"template_id,omitempty" validate:"omitempty,uuid"`
	From       *string `json:"from,omitempty" validate:"omitempty,email"`
	Subject    *string `json:"subject,omitempty"`
	HTML       *string `json:"html,omitempty"`
	Text       *string `json:"text,omitempty"`
}

type UpdateBroadcastRequest struct {
	Name       *string `json:"name,omitempty"`
	AudienceID *string `json:"audience_id,omitempty" validate:"omitempty,uuid"`
	SegmentID  *string `json:"segment_id,omitempty" validate:"omitempty,uuid"`
	From       *string `json:"from,omitempty" validate:"omitempty,email"`
	Subject    *string `json:"subject,omitempty"`
	HTML       *string `json:"html,omitempty"`
	Text       *string `json:"text,omitempty"`
}

type BroadcastResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	AudienceID   *string `json:"audience_id,omitempty"`
	AudienceName *string `json:"audience_name,omitempty"`
	Status       string  `json:"status"`
	Recipients   int     `json:"recipients"`
	Sent         int     `json:"sent"`
	CreatedAt    string  `json:"created_at"`
}
