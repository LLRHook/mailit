package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	tmock "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func TestAuthService_Register_HappyPath(t *testing.T) {
	userRepo := new(tmock.MockUserRepository)
	teamRepo := new(tmock.MockTeamRepository)
	memberRepo := new(tmock.MockTeamMemberRepository)

	svc := NewAuthService(userRepo, teamRepo, memberRepo, "test-secret", time.Hour, bcrypt.MinCost)
	ctx := context.Background()

	// User does not exist yet.
	userRepo.On("GetByEmail", ctx, "alice@example.com").Return(nil, postgres.ErrNotFound)
	userRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil)
	teamRepo.On("Create", ctx, mock.AnythingOfType("*model.Team")).Return(nil)
	memberRepo.On("Create", ctx, mock.AnythingOfType("*model.TeamMember")).Return(nil)

	req := &dto.RegisterRequest{
		Email:    "alice@example.com",
		Password: "securepassword",
		Name:     "Alice",
		TeamName: "My Team",
	}

	resp, err := svc.Register(ctx, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "alice@example.com", resp.User.Email)
	assert.Equal(t, "Alice", resp.User.Name)
	assert.NotEmpty(t, resp.User.ID)

	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	userRepo := new(tmock.MockUserRepository)
	teamRepo := new(tmock.MockTeamRepository)
	memberRepo := new(tmock.MockTeamMemberRepository)

	svc := NewAuthService(userRepo, teamRepo, memberRepo, "test-secret", time.Hour, bcrypt.MinCost)
	ctx := context.Background()

	existingUser := &model.User{Email: "alice@example.com"}
	userRepo.On("GetByEmail", ctx, "alice@example.com").Return(existingUser, nil)

	req := &dto.RegisterRequest{
		Email:    "alice@example.com",
		Password: "securepassword",
		Name:     "Alice",
		TeamName: "My Team",
	}

	resp, err := svc.Register(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_HappyPath(t *testing.T) {
	userRepo := new(tmock.MockUserRepository)
	teamRepo := new(tmock.MockTeamRepository)
	memberRepo := new(tmock.MockTeamMemberRepository)

	svc := NewAuthService(userRepo, teamRepo, memberRepo, "test-secret", time.Hour, bcrypt.MinCost)
	ctx := context.Background()

	// Hash the password so bcrypt compare succeeds.
	hash, _ := bcrypt.GenerateFromPassword([]byte("securepassword"), bcrypt.MinCost)
	user := &model.User{
		Email:        "alice@example.com",
		PasswordHash: string(hash),
		Name:         "Alice",
	}
	user.ID = [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	userRepo.On("GetByEmail", ctx, "alice@example.com").Return(user, nil)

	teamMember := model.TeamMember{
		TeamID: [16]byte{10, 20, 30, 40, 50, 60, 70, 80, 90, 0, 0, 0, 0, 0, 0, 1},
		UserID: user.ID,
		Role:   model.RoleOwner,
	}
	memberRepo.On("ListByUserID", ctx, user.ID).Return([]model.TeamMember{teamMember}, nil)

	req := &dto.LoginRequest{
		Email:    "alice@example.com",
		Password: "securepassword",
	}

	resp, err := svc.Login(ctx, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "alice@example.com", resp.User.Email)
	assert.Equal(t, "Alice", resp.User.Name)

	userRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	userRepo := new(tmock.MockUserRepository)
	teamRepo := new(tmock.MockTeamRepository)
	memberRepo := new(tmock.MockTeamMemberRepository)

	svc := NewAuthService(userRepo, teamRepo, memberRepo, "test-secret", time.Hour, bcrypt.MinCost)
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	user := &model.User{
		Email:        "alice@example.com",
		PasswordHash: string(hash),
		Name:         "Alice",
	}

	userRepo.On("GetByEmail", ctx, "alice@example.com").Return(user, nil)

	req := &dto.LoginRequest{
		Email:    "alice@example.com",
		Password: "wrongpassword",
	}

	resp, err := svc.Login(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email or password")

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := new(tmock.MockUserRepository)
	teamRepo := new(tmock.MockTeamRepository)
	memberRepo := new(tmock.MockTeamMemberRepository)

	svc := NewAuthService(userRepo, teamRepo, memberRepo, "test-secret", time.Hour, bcrypt.MinCost)
	ctx := context.Background()

	userRepo.On("GetByEmail", ctx, "nobody@example.com").Return(nil, fmt.Errorf("not found"))

	req := &dto.LoginRequest{
		Email:    "nobody@example.com",
		Password: "password123",
	}

	resp, err := svc.Login(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email or password")

	userRepo.AssertExpectations(t)
}
