package dto

type LogResponse struct {
	ID        string `json:"id"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	EmailID   string `json:"email_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

type LogListParams struct {
	PaginationParams
	Level string `json:"level,omitempty"`
}
