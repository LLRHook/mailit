package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// APIKeyService defines operations for managing API keys.
type APIKeyService interface {
	Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateAPIKeyRequest) (*dto.APIKeyResponse, error)
	List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.APIKeyResponse], error)
	Delete(ctx context.Context, teamID uuid.UUID, apiKeyID uuid.UUID) error
}

type apiKeyService struct {
	apiKeyRepo   postgres.APIKeyRepository
	apiKeyPrefix string
}

// NewAPIKeyService creates a new APIKeyService.
func NewAPIKeyService(apiKeyRepo postgres.APIKeyRepository, apiKeyPrefix string) APIKeyService {
	return &apiKeyService{
		apiKeyRepo:   apiKeyRepo,
		apiKeyPrefix: apiKeyPrefix,
	}
}

func (s *apiKeyService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateAPIKeyRequest) (*dto.APIKeyResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	// Default permission to "full" if not specified.
	permission := req.Permission
	if permission == "" {
		permission = model.PermissionFull
	}

	// Generate the API key.
	plaintext, hash, keyPrefix, err := pkg.GenerateAPIKey(s.apiKeyPrefix)
	if err != nil {
		return nil, fmt.Errorf("generating API key: %w", err)
	}

	// Parse optional domain ID.
	var domainID *uuid.UUID
	if req.DomainID != nil && *req.DomainID != "" {
		id, err := uuid.Parse(*req.DomainID)
		if err != nil {
			return nil, fmt.Errorf("invalid domain_id: %w", err)
		}
		domainID = &id
	}

	now := time.Now().UTC()

	apiKey := &model.APIKey{
		ID:         uuid.New(),
		TeamID:     teamID,
		Name:       req.Name,
		KeyHash:    hash,
		KeyPrefix:  keyPrefix,
		Permission: permission,
		DomainID:   domainID,
		CreatedAt:  now,
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("creating API key: %w", err)
	}

	// Return the plaintext token only on creation.
	return &dto.APIKeyResponse{
		ID:         apiKey.ID.String(),
		Name:       apiKey.Name,
		Token:      plaintext,
		KeyPrefix:  apiKey.KeyPrefix,
		Permission: apiKey.Permission,
		CreatedAt:  apiKey.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *apiKeyService) List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.APIKeyResponse], error) {
	keys, err := s.apiKeyRepo.ListByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing API keys: %w", err)
	}

	responses := make([]dto.APIKeyResponse, 0, len(keys))
	for _, k := range keys {
		responses = append(responses, dto.APIKeyResponse{
			ID:         k.ID.String(),
			Name:       k.Name,
			KeyPrefix:  k.KeyPrefix,
			Permission: k.Permission,
			CreatedAt:  k.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.ListResponse[dto.APIKeyResponse]{Data: responses}, nil
}

func (s *apiKeyService) Delete(ctx context.Context, teamID uuid.UUID, apiKeyID uuid.UUID) error {
	// Verify the key belongs to the team by listing and checking.
	keys, err := s.apiKeyRepo.ListByTeamID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("listing API keys: %w", err)
	}

	found := false
	for _, k := range keys {
		if k.ID == apiKeyID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("API key not found")
	}

	if err := s.apiKeyRepo.Delete(ctx, apiKeyID); err != nil {
		return fmt.Errorf("deleting API key: %w", err)
	}

	return nil
}
