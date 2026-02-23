package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/server/middleware"
)

// AuthService defines authentication and registration operations.
type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
}

type authService struct {
	userRepo       postgres.UserRepository
	teamRepo       postgres.TeamRepository
	teamMemberRepo postgres.TeamMemberRepository
	jwtSecret      string
	jwtExpiry      time.Duration
	bcryptCost     int
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo postgres.UserRepository,
	teamRepo postgres.TeamRepository,
	teamMemberRepo postgres.TeamMemberRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
	bcryptCost int,
) AuthService {
	return &authService{
		userRepo:       userRepo,
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		jwtSecret:      jwtSecret,
		jwtExpiry:      jwtExpiry,
		bcryptCost:     bcryptCost,
	}
}

// nonAlphanumHyphen matches characters that are not lowercase alphanumeric or hyphens.
var nonAlphanumHyphen = regexp.MustCompile(`[^a-z0-9-]`)

// slugify converts a team name to a URL-safe slug.
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	s = nonAlphanumHyphen.ReplaceAllString(s, "")
	// Collapse consecutive hyphens.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if s == "" {
		s = "team"
	}
	return s
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	// Check if user already exists.
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, fmt.Errorf("a user with this email already exists")
	}

	// Hash password.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now().UTC()

	// Create user.
	user := &model.User{
		ID:            uuid.New(),
		Email:         req.Email,
		PasswordHash:  string(hash),
		Name:          req.Name,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Create team with slug derived from team name.
	team := &model.Team{
		ID:        uuid.New(),
		Name:      req.TeamName,
		Slug:      slugify(req.TeamName),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("creating team: %w", err)
	}

	// Create team member with owner role.
	member := &model.TeamMember{
		ID:        uuid.New(),
		TeamID:    team.ID,
		UserID:    user.ID,
		Role:      model.RoleOwner,
		CreatedAt: now,
	}
	if err := s.teamMemberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("creating team member: %w", err)
	}

	// Generate JWT.
	token, err := middleware.GenerateJWT(s.jwtSecret, user.ID, team.ID, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	resp := &dto.AuthResponse{
		Token: token,
	}
	resp.User.ID = user.ID.String()
	resp.User.Email = user.Email
	resp.User.Name = user.Name

	return resp, nil
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	// Find user by email.
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Compare password.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Find the user's first team.
	members, err := s.teamMemberRepo.ListByUserID(ctx, user.ID)
	if err != nil || len(members) == 0 {
		return nil, fmt.Errorf("user has no team memberships")
	}
	teamID := members[0].TeamID

	// Generate JWT.
	token, err := middleware.GenerateJWT(s.jwtSecret, user.ID, teamID, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	resp := &dto.AuthResponse{
		Token: token,
	}
	resp.User.ID = user.ID.String()
	resp.User.Email = user.Email
	resp.User.Name = user.Name

	return resp, nil
}
