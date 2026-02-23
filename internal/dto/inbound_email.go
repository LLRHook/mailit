package dto

type InboundEmailResponse struct {
	ID        string   `json:"id"`
	From      string   `json:"from"`
	To        []string `json:"to"`
	Subject   string   `json:"subject"`
	HTML      *string  `json:"html,omitempty"`
	Text      *string  `json:"text,omitempty"`
	CreatedAt string   `json:"created_at"`
}
