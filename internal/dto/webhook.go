package dto

type CreateWebhookRequest struct {
	URL    string   `json:"url" validate:"required,url"`
	Events []string `json:"events" validate:"required,min=1"`
}

type UpdateWebhookRequest struct {
	URL    *string  `json:"url,omitempty" validate:"omitempty,url"`
	Events []string `json:"events,omitempty"`
	Active *bool    `json:"active,omitempty"`
}

type WebhookResponse struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"created_at"`
}
