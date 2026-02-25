package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/worker"
)

const (
	// idempotencyKeyPrefix is the Redis key prefix for idempotency keys.
	idempotencyKeyPrefix = "idempotency:"

	// idempotencyKeyTTL is how long an idempotency key is remembered.
	idempotencyKeyTTL = 24 * time.Hour
)

// EmailService defines operations for transactional email sending.
type EmailService interface {
	Send(ctx context.Context, teamID uuid.UUID, req *dto.SendEmailRequest) (*dto.SendEmailResponse, error)
	BatchSend(ctx context.Context, teamID uuid.UUID, req *dto.BatchSendEmailRequest) (*dto.BatchSendEmailResponse, error)
	List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.EmailResponse], error)
	Get(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID) (*dto.EmailResponse, error)
	Update(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID, req map[string]interface{}) (*dto.EmailResponse, error)
	Cancel(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID) (*dto.EmailResponse, error)
}

type emailService struct {
	emailRepo       postgres.EmailRepository
	suppressionRepo postgres.SuppressionRepository
	asynqClient     *asynq.Client
	redisClient     *redis.Client
}

// NewEmailService creates a new EmailService.
func NewEmailService(
	emailRepo postgres.EmailRepository,
	suppressionRepo postgres.SuppressionRepository,
	asynqClient *asynq.Client,
	redisClient *redis.Client,
) EmailService {
	return &emailService{
		emailRepo:       emailRepo,
		suppressionRepo: suppressionRepo,
		asynqClient:     asynqClient,
		redisClient:     redisClient,
	}
}

func (s *emailService) Send(ctx context.Context, teamID uuid.UUID, req *dto.SendEmailRequest) (*dto.SendEmailResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	// Check idempotency key to prevent duplicate sends.
	if req.IdempotencyKey != nil && *req.IdempotencyKey != "" {
		key := idempotencyKeyPrefix + teamID.String() + ":" + *req.IdempotencyKey
		existing, err := s.redisClient.Get(ctx, key).Result()
		if err == nil && existing != "" {
			// Return the previously created email ID.
			return &dto.SendEmailResponse{ID: existing}, nil
		}
	}

	// Check suppression list for each recipient.
	for _, addr := range req.To {
		entry, err := s.suppressionRepo.GetByTeamAndEmail(ctx, teamID, addr)
		if err != nil && !errors.Is(err, postgres.ErrNotFound) {
			return nil, fmt.Errorf("checking suppression list: %w", err)
		}
		if entry != nil {
			return nil, fmt.Errorf("recipient %s is on the suppression list", addr)
		}
	}

	now := time.Now().UTC()

	// Determine initial status.
	status := model.EmailStatusQueued
	var scheduledAt *time.Time
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			return nil, fmt.Errorf("invalid scheduled_at format: %w", err)
		}
		scheduledAt = &t
		status = model.EmailStatusScheduled
	}

	// Convert tags to string slice.
	tags := make([]string, 0, len(req.Tags))
	for _, tag := range req.Tags {
		tags = append(tags, tag.Name+":"+tag.Value)
	}

	// Convert headers to JSONMap.
	headers := make(model.JSONMap)
	for k, v := range req.Headers {
		headers[k] = v
	}

	// Convert attachments to JSONArray.
	attachments := make(model.JSONArray, 0, len(req.Attachments))
	for _, a := range req.Attachments {
		attachments = append(attachments, map[string]interface{}{
			"filename":     a.Filename,
			"content":      a.Content,
			"content_type": a.ContentType,
		})
	}

	email := &model.Email{
		ID:             uuid.New(),
		TeamID:         teamID,
		FromAddress:    req.From,
		ToAddresses:    req.To,
		CcAddresses:    req.Cc,
		BccAddresses:   req.Bcc,
		ReplyTo:        req.ReplyTo,
		Subject:        req.Subject,
		HTMLBody:       req.HTML,
		TextBody:       req.Text,
		Status:         status,
		ScheduledAt:    scheduledAt,
		Tags:           tags,
		Headers:        headers,
		Attachments:    attachments,
		IdempotencyKey: req.IdempotencyKey,
		RetryCount:     0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.emailRepo.Create(ctx, email); err != nil {
		return nil, fmt.Errorf("creating email record: %w", err)
	}

	// Store idempotency key in Redis.
	if req.IdempotencyKey != nil && *req.IdempotencyKey != "" {
		key := idempotencyKeyPrefix + teamID.String() + ":" + *req.IdempotencyKey
		s.redisClient.Set(ctx, key, email.ID.String(), idempotencyKeyTTL)
	}

	// Enqueue the send task.
	payload, err := json.Marshal(map[string]string{
		"email_id": email.ID.String(),
		"team_id":  teamID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshalling task payload: %w", err)
	}

	task := asynq.NewTask(worker.TaskEmailSend, payload)
	opts := []asynq.Option{
		asynq.Queue(worker.QueueCritical),
		asynq.MaxRetry(3),
	}
	if scheduledAt != nil {
		opts = append(opts, asynq.ProcessAt(*scheduledAt))
	}

	if _, err := s.asynqClient.Enqueue(task, opts...); err != nil {
		return nil, fmt.Errorf("enqueueing send task: %w", err)
	}

	return &dto.SendEmailResponse{ID: email.ID.String()}, nil
}

func (s *emailService) BatchSend(ctx context.Context, teamID uuid.UUID, req *dto.BatchSendEmailRequest) (*dto.BatchSendEmailResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	responses := make([]dto.SendEmailResponse, 0, len(req.Emails))
	for i := range req.Emails {
		resp, err := s.Send(ctx, teamID, &req.Emails[i])
		if err != nil {
			return nil, fmt.Errorf("sending email %d: %w", i, err)
		}
		responses = append(responses, *resp)
	}

	return &dto.BatchSendEmailResponse{Data: responses}, nil
}

func (s *emailService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.EmailResponse], error) {
	params.Normalize()

	emails, total, err := s.emailRepo.List(ctx, teamID, params.PerPage, params.Offset())
	if err != nil {
		return nil, fmt.Errorf("listing emails: %w", err)
	}

	data := make([]dto.EmailResponse, 0, len(emails))
	for _, e := range emails {
		data = append(data, emailToResponse(&e))
	}

	totalPages := 0
	if params.PerPage > 0 {
		totalPages = (total + params.PerPage - 1) / params.PerPage
	}

	return &dto.PaginatedResponse[dto.EmailResponse]{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
		HasMore:    params.Page < totalPages,
	}, nil
}

func (s *emailService) Get(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID) (*dto.EmailResponse, error) {
	email, err := s.emailRepo.GetByID(ctx, emailID)
	if err != nil {
		return nil, fmt.Errorf("email not found: %w", err)
	}

	// Verify the email belongs to the team.
	if email.TeamID != teamID {
		return nil, fmt.Errorf("email not found: %w", postgres.ErrNotFound)
	}

	resp := emailToResponse(email)
	return &resp, nil
}

func (s *emailService) Update(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID, req map[string]interface{}) (*dto.EmailResponse, error) {
	email, err := s.emailRepo.GetByID(ctx, emailID)
	if err != nil {
		return nil, fmt.Errorf("email not found: %w", err)
	}

	// Verify the email belongs to the team.
	if email.TeamID != teamID {
		return nil, fmt.Errorf("email not found: %w", postgres.ErrNotFound)
	}

	// Only scheduled emails can be updated.
	if email.Status != model.EmailStatusScheduled {
		return nil, fmt.Errorf("only scheduled emails can be updated")
	}

	if scheduledAt, ok := req["scheduled_at"].(string); ok {
		t, err := time.Parse(time.RFC3339, scheduledAt)
		if err != nil {
			return nil, fmt.Errorf("invalid scheduled_at format: %w", err)
		}
		email.ScheduledAt = &t
	}

	email.UpdatedAt = time.Now().UTC()

	if err := s.emailRepo.Update(ctx, email); err != nil {
		return nil, fmt.Errorf("updating email: %w", err)
	}

	resp := emailToResponse(email)
	return &resp, nil
}

func (s *emailService) Cancel(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID) (*dto.EmailResponse, error) {
	email, err := s.emailRepo.GetByID(ctx, emailID)
	if err != nil {
		return nil, fmt.Errorf("email not found: %w", err)
	}

	// Verify the email belongs to the team.
	if email.TeamID != teamID {
		return nil, fmt.Errorf("email not found: %w", postgres.ErrNotFound)
	}

	// Only queued or scheduled emails can be cancelled.
	if email.Status != model.EmailStatusQueued && email.Status != model.EmailStatusScheduled {
		return nil, fmt.Errorf("only queued or scheduled emails can be cancelled")
	}

	email.Status = model.EmailStatusCancelled
	email.UpdatedAt = time.Now().UTC()

	if err := s.emailRepo.Update(ctx, email); err != nil {
		return nil, fmt.Errorf("cancelling email: %w", err)
	}

	resp := emailToResponse(email)
	return &resp, nil
}

// emailToResponse converts a model.Email to a dto.EmailResponse.
func emailToResponse(e *model.Email) dto.EmailResponse {
	resp := dto.EmailResponse{
		ID:        e.ID.String(),
		From:      e.FromAddress,
		To:        e.ToAddresses,
		Cc:        e.CcAddresses,
		Bcc:       e.BccAddresses,
		ReplyTo:   e.ReplyTo,
		Subject:   e.Subject,
		HTML:      e.HTMLBody,
		Text:      e.TextBody,
		Status:    e.Status,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		LastEvent: e.Status,
	}

	if e.ScheduledAt != nil {
		s := e.ScheduledAt.Format(time.RFC3339)
		resp.ScheduledAt = &s
	}
	if e.SentAt != nil {
		s := e.SentAt.Format(time.RFC3339)
		resp.SentAt = &s
	}

	return resp
}
