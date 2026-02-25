package handler

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/server/middleware"
	"github.com/mailit-dev/mailit/internal/worker"
)

type ContactImportHandler struct {
	importJobRepo postgres.ContactImportJobRepository
	audienceRepo  postgres.AudienceRepository
	asynqClient   *asynq.Client
}

func NewContactImportHandler(
	importJobRepo postgres.ContactImportJobRepository,
	audienceRepo postgres.AudienceRepository,
	asynqClient *asynq.Client,
) *ContactImportHandler {
	return &ContactImportHandler{
		importJobRepo: importJobRepo,
		audienceRepo:  audienceRepo,
		asynqClient:   asynqClient,
	}
}

// Import handles POST /audiences/{audienceId}/contacts/import.
func (h *ContactImportHandler) Import(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	audienceID, err := uuid.Parse(chi.URLParam(r, "audienceId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid audience id")
		return
	}

	// Verify audience belongs to team.
	audience, err := h.audienceRepo.GetByID(r.Context(), audienceID)
	if err != nil || audience.TeamID != auth.TeamID {
		pkg.Error(w, http.StatusNotFound, "audience not found")
		return
	}

	// Parse multipart form (max 10MB).
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "missing 'file' field")
		return
	}
	defer func() { _ = file.Close() }()

	// Read CSV content.
	csvBytes, err := io.ReadAll(file)
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "failed to read file")
		return
	}

	csvData := string(csvBytes)

	// Quick validation: count rows.
	reader := csv.NewReader(strings.NewReader(csvData))
	rowCount := 0
	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			pkg.Error(w, http.StatusBadRequest, fmt.Sprintf("invalid CSV at row %d: %v", rowCount+1, err))
			return
		}
		rowCount++
	}

	if rowCount < 2 {
		pkg.Error(w, http.StatusBadRequest, "CSV must have a header row and at least one data row")
		return
	}

	dataRows := rowCount - 1 // Exclude header.

	// Create the import job.
	now := time.Now().UTC()
	job := &model.ContactImportJob{
		ID:         uuid.New(),
		TeamID:     auth.TeamID,
		AudienceID: audienceID,
		Status:     model.ImportStatusPending,
		TotalRows:  dataRows,
		CSVData:    csvData,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.importJobRepo.Create(r.Context(), job); err != nil {
		pkg.Error(w, http.StatusInternalServerError, "failed to create import job")
		return
	}

	// Enqueue the import task.
	task, err := worker.NewContactImportTask(job.ID, auth.TeamID)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, "failed to create import task")
		return
	}

	if _, err := h.asynqClient.Enqueue(task); err != nil {
		pkg.Error(w, http.StatusInternalServerError, "failed to enqueue import task")
		return
	}

	pkg.JSON(w, http.StatusAccepted, map[string]interface{}{
		"job_id":     job.ID.String(),
		"status":     job.Status,
		"total_rows": job.TotalRows,
	})
}

// GetImportStatus handles GET /audiences/{audienceId}/contacts/import/{jobId}.
func (h *ContactImportHandler) GetImportStatus(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobID, err := uuid.Parse(chi.URLParam(r, "jobId"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.importJobRepo.GetByID(r.Context(), jobID)
	if err != nil {
		pkg.Error(w, http.StatusNotFound, "import job not found")
		return
	}

	if job.TeamID != auth.TeamID {
		pkg.Error(w, http.StatusNotFound, "import job not found")
		return
	}

	pkg.JSON(w, http.StatusOK, map[string]interface{}{
		"id":             job.ID.String(),
		"status":         job.Status,
		"total_rows":     job.TotalRows,
		"processed_rows": job.ProcessedRows,
		"created_rows":   job.CreatedRows,
		"updated_rows":   job.UpdatedRows,
		"skipped_rows":   job.SkippedRows,
		"failed_rows":    job.FailedRows,
		"error":          job.Error,
		"created_at":     job.CreatedAt.Format(time.RFC3339),
	})
}
