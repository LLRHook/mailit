package worker

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Task type constants for all background jobs.
const (
	TaskEmailSend      = "email:send"
	TaskEmailBatchSend = "email:send_batch"
	TaskBroadcastSend  = "broadcast:send"
	TaskDomainVerify   = "domain:verify"
	TaskWebhookDeliver = "webhook:deliver"
	TaskBounceProcess  = "bounce:process"
	TaskInboundProcess = "inbound:process"
	TaskCleanupExpired    = "cleanup:expired"
	TaskMetricsAggregate  = "metrics:aggregate"
	TaskContactImport     = "contact:import"
)

// Queue names and their intended priority levels.
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

// EmailSendPayload is the payload for sending a single email.
type EmailSendPayload struct {
	EmailID uuid.UUID `json:"email_id"`
	TeamID  uuid.UUID `json:"team_id"`
}

// EmailBatchSendPayload is the payload for sending a batch of emails.
type EmailBatchSendPayload struct {
	EmailIDs []uuid.UUID `json:"email_ids"`
	TeamID   uuid.UUID   `json:"team_id"`
}

// BroadcastSendPayload is the payload for sending a broadcast.
type BroadcastSendPayload struct {
	BroadcastID uuid.UUID `json:"broadcast_id"`
	TeamID      uuid.UUID `json:"team_id"`
}

// DomainVerifyPayload is the payload for verifying a domain's DNS records.
type DomainVerifyPayload struct {
	DomainID uuid.UUID `json:"domain_id"`
	TeamID   uuid.UUID `json:"team_id"`
}

// WebhookDeliverPayload is the payload for delivering a webhook event.
type WebhookDeliverPayload struct {
	WebhookEventID uuid.UUID `json:"webhook_event_id"`
}

// BounceProcessPayload is the payload for processing a bounce.
type BounceProcessPayload struct {
	EmailID   uuid.UUID `json:"email_id"`
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Recipient string    `json:"recipient"`
}

// InboundProcessPayload is the payload for processing an inbound email.
type InboundProcessPayload struct {
	InboundEmailID uuid.UUID `json:"inbound_email_id"`
}

// NewEmailSendTask creates an asynq task for sending a single email.
func NewEmailSendTask(emailID, teamID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailSendPayload{EmailID: emailID, TeamID: teamID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskEmailSend, payload, asynq.Queue(QueueCritical), asynq.MaxRetry(8)), nil
}

// NewEmailBatchSendTask creates an asynq task for sending a batch of emails.
func NewEmailBatchSendTask(emailIDs []uuid.UUID, teamID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailBatchSendPayload{EmailIDs: emailIDs, TeamID: teamID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskEmailBatchSend, payload, asynq.Queue(QueueCritical), asynq.MaxRetry(8)), nil
}

// NewBroadcastSendTask creates an asynq task for sending a broadcast.
func NewBroadcastSendTask(broadcastID, teamID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(BroadcastSendPayload{BroadcastID: broadcastID, TeamID: teamID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskBroadcastSend, payload, asynq.Queue(QueueCritical), asynq.MaxRetry(3)), nil
}

// NewDomainVerifyTask creates an asynq task for verifying a domain's DNS records.
func NewDomainVerifyTask(domainID, teamID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(DomainVerifyPayload{DomainID: domainID, TeamID: teamID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskDomainVerify, payload, asynq.Queue(QueueDefault), asynq.MaxRetry(3)), nil
}

// NewWebhookDeliverTask creates an asynq task for delivering a webhook event.
func NewWebhookDeliverTask(webhookEventID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(WebhookDeliverPayload{WebhookEventID: webhookEventID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskWebhookDeliver, payload, asynq.Queue(QueueDefault), asynq.MaxRetry(5)), nil
}

// NewBounceProcessTask creates an asynq task for processing a bounce notification.
func NewBounceProcessTask(emailID uuid.UUID, code int, message, recipient string) (*asynq.Task, error) {
	payload, err := json.Marshal(BounceProcessPayload{
		EmailID:   emailID,
		Code:      code,
		Message:   message,
		Recipient: recipient,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskBounceProcess, payload, asynq.Queue(QueueDefault), asynq.MaxRetry(3)), nil
}

// NewInboundProcessTask creates an asynq task for processing an inbound email.
func NewInboundProcessTask(inboundEmailID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(InboundProcessPayload{InboundEmailID: inboundEmailID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskInboundProcess, payload, asynq.Queue(QueueDefault), asynq.MaxRetry(3)), nil
}

// ContactImportPayload is the payload for importing contacts from CSV.
type ContactImportPayload struct {
	JobID  uuid.UUID `json:"job_id"`
	TeamID uuid.UUID `json:"team_id"`
}

// NewContactImportTask creates an asynq task for importing contacts from CSV.
func NewContactImportTask(jobID, teamID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(ContactImportPayload{JobID: jobID, TeamID: teamID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskContactImport, payload, asynq.Queue(QueueDefault), asynq.MaxRetry(1)), nil
}

// NewCleanupExpiredTask creates an asynq task for cleaning up expired data.
func NewCleanupExpiredTask() (*asynq.Task, error) {
	return asynq.NewTask(TaskCleanupExpired, nil, asynq.Queue(QueueLow), asynq.MaxRetry(1)), nil
}

// NewMetricsAggregateTask creates an asynq task for aggregating email metrics.
func NewMetricsAggregateTask() (*asynq.Task, error) {
	return asynq.NewTask(TaskMetricsAggregate, nil, asynq.Queue(QueueLow), asynq.MaxRetry(1)), nil
}
