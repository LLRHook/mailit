package worker

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// ContactImportHandler processes contact:import tasks.
type ContactImportHandler struct {
	importJobRepo postgres.ContactImportJobRepository
	contactRepo   postgres.ContactRepository
	logger        *slog.Logger
}

// NewContactImportHandler creates a new ContactImportHandler.
func NewContactImportHandler(
	importJobRepo postgres.ContactImportJobRepository,
	contactRepo postgres.ContactRepository,
	logger *slog.Logger,
) *ContactImportHandler {
	return &ContactImportHandler{
		importJobRepo: importJobRepo,
		contactRepo:   contactRepo,
		logger:        logger,
	}
}

// ProcessTask handles the contact:import task.
func (h *ContactImportHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p ContactImportPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshalling contact:import payload: %w", err)
	}

	log := h.logger.With("job_id", p.JobID, "team_id", p.TeamID)

	// 1. Get the import job from DB.
	job, err := h.importJobRepo.GetByID(ctx, p.JobID)
	if err != nil {
		return fmt.Errorf("fetching import job %s: %w", p.JobID, err)
	}

	if job.Status != model.ImportStatusPending {
		log.Info("skipping import job with non-pending status", "status", job.Status)
		return nil
	}

	// 2. Mark as processing.
	job.Status = model.ImportStatusProcessing
	job.UpdatedAt = time.Now().UTC()
	if err := h.importJobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("updating import job to processing: %w", err)
	}

	// 3. Parse CSV.
	reader := csv.NewReader(strings.NewReader(job.CSVData))

	// Read header row.
	header, err := reader.Read()
	if err != nil {
		return h.failJob(ctx, job, fmt.Sprintf("reading CSV header: %v", err))
	}

	// Map column names to indices.
	colMap := make(map[string]int, len(header))
	for i, col := range header {
		colMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	emailCol, ok := colMap["email"]
	if !ok {
		return h.failJob(ctx, job, "CSV must have an 'email' column")
	}

	firstNameCol, hasFirstName := colMap["first_name"]
	lastNameCol, hasLastName := colMap["last_name"]

	// 4. Process rows.
	now := time.Now().UTC()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			job.FailedRows++
			job.ProcessedRows++
			continue
		}

		job.ProcessedRows++

		if emailCol >= len(record) {
			job.FailedRows++
			continue
		}

		email := strings.TrimSpace(record[emailCol])
		if email == "" {
			job.SkippedRows++
			continue
		}

		var firstName, lastName *string
		if hasFirstName && firstNameCol < len(record) {
			v := strings.TrimSpace(record[firstNameCol])
			if v != "" {
				firstName = &v
			}
		}
		if hasLastName && lastNameCol < len(record) {
			v := strings.TrimSpace(record[lastNameCol])
			if v != "" {
				lastName = &v
			}
		}

		// Check for existing contact.
		existing, _ := h.contactRepo.GetByAudienceAndEmail(ctx, job.AudienceID, email)
		if existing != nil {
			// Update existing contact if names provided.
			updated := false
			if firstName != nil {
				existing.FirstName = firstName
				updated = true
			}
			if lastName != nil {
				existing.LastName = lastName
				updated = true
			}
			if updated {
				existing.UpdatedAt = now
				if updateErr := h.contactRepo.Update(ctx, existing); updateErr != nil {
					log.Error("failed to update contact during import", "email", email, "error", updateErr)
					job.FailedRows++
					continue
				}
				job.UpdatedRows++
			} else {
				job.SkippedRows++
			}
			continue
		}

		// Create new contact.
		contact := &model.Contact{
			ID:         uuid.New(),
			AudienceID: job.AudienceID,
			Email:      email,
			FirstName:  firstName,
			LastName:   lastName,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		if createErr := h.contactRepo.Create(ctx, contact); createErr != nil {
			log.Error("failed to create contact during import", "email", email, "error", createErr)
			job.FailedRows++
			continue
		}

		job.CreatedRows++

		// Periodically update progress.
		if job.ProcessedRows%100 == 0 {
			job.UpdatedAt = time.Now().UTC()
			_ = h.importJobRepo.Update(ctx, job)
		}
	}

	// 5. Mark as completed.
	job.Status = model.ImportStatusCompleted
	job.UpdatedAt = time.Now().UTC()
	if err := h.importJobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("updating import job to completed: %w", err)
	}

	log.Info("contact import completed",
		"total", job.TotalRows,
		"created", job.CreatedRows,
		"updated", job.UpdatedRows,
		"skipped", job.SkippedRows,
		"failed", job.FailedRows,
	)
	return nil
}

// failJob marks an import job as failed with the given error.
func (h *ContactImportHandler) failJob(ctx context.Context, job *model.ContactImportJob, errMsg string) error {
	job.Status = model.ImportStatusFailed
	job.Error = &errMsg
	job.UpdatedAt = time.Now().UTC()
	if err := h.importJobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("marking import job as failed: %w", err)
	}
	return fmt.Errorf("import failed: %s", errMsg)
}
