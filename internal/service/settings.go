package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/server/middleware"
)

// SettingsService defines operations for the settings page.
type SettingsService interface {
	GetUsage(ctx context.Context, teamID uuid.UUID) (*dto.UsageResponse, error)
	GetTeam(ctx context.Context, teamID uuid.UUID) (*dto.TeamResponse, error)
	UpdateTeam(ctx context.Context, teamID uuid.UUID, req *dto.UpdateTeamRequest) error
	GetSMTPConfig() *dto.SMTPConfigResponse
	InviteMember(ctx context.Context, teamID uuid.UUID, req *dto.InviteMemberRequest) (*model.TeamInvitation, error)
	AcceptInvite(ctx context.Context, req *dto.AcceptInviteRequest) (*dto.AuthResponse, error)
}

// SMTPDisplayConfig holds the SMTP settings to display to users.
type SMTPDisplayConfig struct {
	Host       string
	Port       int
	Encryption string
}

type settingsService struct {
	settingsRepo   postgres.SettingsRepository
	invitationRepo postgres.TeamInvitationRepository
	userRepo       postgres.UserRepository
	teamMemberRepo postgres.TeamMemberRepository
	smtpConfig     SMTPDisplayConfig
	jwtSecret      string
	jwtExpiry      time.Duration
	bcryptCost     int
}

// NewSettingsService creates a new SettingsService.
func NewSettingsService(
	settingsRepo postgres.SettingsRepository,
	invitationRepo postgres.TeamInvitationRepository,
	userRepo postgres.UserRepository,
	teamMemberRepo postgres.TeamMemberRepository,
	smtpConfig SMTPDisplayConfig,
	jwtSecret string,
	jwtExpiry time.Duration,
	bcryptCost int,
) SettingsService {
	return &settingsService{
		settingsRepo:   settingsRepo,
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		teamMemberRepo: teamMemberRepo,
		smtpConfig:     smtpConfig,
		jwtSecret:      jwtSecret,
		jwtExpiry:      jwtExpiry,
		bcryptCost:     bcryptCost,
	}
}

func (s *settingsService) GetUsage(ctx context.Context, teamID uuid.UUID) (*dto.UsageResponse, error) {
	usage, err := s.settingsRepo.GetUsageCounts(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("getting usage counts: %w", err)
	}
	return usage, nil
}

func (s *settingsService) GetTeam(ctx context.Context, teamID uuid.UUID) (*dto.TeamResponse, error) {
	team, err := s.settingsRepo.GetTeamWithMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("getting team: %w", err)
	}
	return team, nil
}

func (s *settingsService) UpdateTeam(ctx context.Context, teamID uuid.UUID, req *dto.UpdateTeamRequest) error {
	if err := s.settingsRepo.UpdateTeamName(ctx, teamID, req.Name); err != nil {
		return fmt.Errorf("updating team name: %w", err)
	}
	return nil
}

func (s *settingsService) GetSMTPConfig() *dto.SMTPConfigResponse {
	return &dto.SMTPConfigResponse{
		Host:       s.smtpConfig.Host,
		Port:       s.smtpConfig.Port,
		Username:   "your-api-key",
		Password:   "Use any API key as the password",
		Encryption: s.smtpConfig.Encryption,
	}
}

func (s *settingsService) InviteMember(ctx context.Context, teamID uuid.UUID, req *dto.InviteMemberRequest) (*model.TeamInvitation, error) {
	// Generate a secure random token.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generating invite token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	invitation := &model.TeamInvitation{
		ID:        uuid.New(),
		TeamID:    teamID,
		Email:     req.Email,
		Role:      req.Role,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour), // 7 day expiry
		CreatedAt: time.Now().UTC(),
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("creating invitation: %w", err)
	}

	return invitation, nil
}

func (s *settingsService) AcceptInvite(ctx context.Context, req *dto.AcceptInviteRequest) (*dto.AuthResponse, error) {
	// Look up the invitation.
	invitation, err := s.invitationRepo.GetByToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, fmt.Errorf("invalid or expired invitation")
		}
		return nil, fmt.Errorf("looking up invitation: %w", err)
	}

	// Check if already accepted.
	if invitation.AcceptedAt != nil {
		return nil, fmt.Errorf("invitation already accepted")
	}

	// Check if expired.
	if time.Now().UTC().After(invitation.ExpiresAt) {
		return nil, fmt.Errorf("invitation has expired")
	}

	// Hash the password.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Create the user.
	user := &model.User{
		ID:            uuid.New(),
		Email:         invitation.Email,
		PasswordHash:  string(hash),
		Name:          req.Name,
		EmailVerified: true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Add as team member.
	member := &model.TeamMember{
		ID:        uuid.New(),
		TeamID:    invitation.TeamID,
		UserID:    user.ID,
		Role:      invitation.Role,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.teamMemberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("adding team member: %w", err)
	}

	// Mark invitation as accepted.
	if err := s.invitationRepo.MarkAccepted(ctx, invitation.ID); err != nil {
		return nil, fmt.Errorf("marking invitation accepted: %w", err)
	}

	// Generate JWT for immediate login.
	token, err := middleware.GenerateJWT(s.jwtSecret, user.ID, invitation.TeamID, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("generating JWT: %w", err)
	}

	resp := &dto.AuthResponse{
		Token: token,
	}
	resp.User.ID = user.ID.String()
	resp.User.Email = user.Email
	resp.User.Name = user.Name

	return resp, nil
}
