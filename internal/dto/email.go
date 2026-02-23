package dto

type SendEmailRequest struct {
	From           string            `json:"from" validate:"required,email"`
	To             []string          `json:"to" validate:"required,min=1,dive,email"`
	Cc             []string          `json:"cc,omitempty" validate:"omitempty,dive,email"`
	Bcc            []string          `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	ReplyTo        *string           `json:"reply_to,omitempty" validate:"omitempty,email"`
	Subject        string            `json:"subject" validate:"required"`
	HTML           *string           `json:"html,omitempty"`
	Text           *string           `json:"text,omitempty"`
	ScheduledAt    *string           `json:"scheduled_at,omitempty"`
	Tags           []Tag             `json:"tags,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	Attachments    []Attachment      `json:"attachments,omitempty"`
	IdempotencyKey *string           `json:"-"` // from header
}

type Tag struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type Attachment struct {
	Filename    string `json:"filename" validate:"required"`
	Content     string `json:"content" validate:"required"` // base64
	ContentType string `json:"content_type,omitempty"`
}

type SendEmailResponse struct {
	ID string `json:"id"`
}

type BatchSendEmailRequest struct {
	Emails []SendEmailRequest `json:"emails" validate:"required,min=1,max=100,dive"`
}

type BatchSendEmailResponse struct {
	Data []SendEmailResponse `json:"data"`
}

type UpdateEmailRequest struct {
	ScheduledAt *string `json:"scheduled_at,omitempty"`
}

type EmailResponse struct {
	ID          string   `json:"id"`
	From        string   `json:"from"`
	To          []string `json:"to"`
	Cc          []string `json:"cc,omitempty"`
	Bcc         []string `json:"bcc,omitempty"`
	ReplyTo     *string  `json:"reply_to,omitempty"`
	Subject     string   `json:"subject"`
	HTML        *string  `json:"html,omitempty"`
	Text        *string  `json:"text,omitempty"`
	Status      string   `json:"status"`
	ScheduledAt *string  `json:"scheduled_at,omitempty"`
	SentAt      *string  `json:"sent_at,omitempty"`
	CreatedAt   string   `json:"created_at"`
	LastEvent   string   `json:"last_event,omitempty"`
}
