package dto

type CreateTemplateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	Subject     *string `json:"subject,omitempty"`
	HTML        *string `json:"html,omitempty"`
	Text        *string `json:"text,omitempty"`
}

type UpdateTemplateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Subject     *string `json:"subject,omitempty"`
	HTML        *string `json:"html,omitempty"`
	Text        *string `json:"text,omitempty"`
}

type TemplateResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	CreatedAt   string  `json:"created_at"`
}
