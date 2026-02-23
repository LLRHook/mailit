package dto

type CreateContactRequest struct {
	Email        string  `json:"email" validate:"required,email"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	Unsubscribed *bool   `json:"unsubscribed,omitempty"`
}

type UpdateContactRequest struct {
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	Unsubscribed *bool   `json:"unsubscribed,omitempty"`
}

type ContactResponse struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	Unsubscribed bool    `json:"unsubscribed"`
	CreatedAt    string  `json:"created_at"`
}
