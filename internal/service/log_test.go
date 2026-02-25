package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/testutil"
	tmock "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func TestLogService_List_WithLevelFilter(t *testing.T) {
	logRepo := new(tmock.MockLogRepository)
	svc := NewLogService(logRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	log1 := model.Log{
		ID:       uuid.New(),
		TeamID:   teamID,
		Level:    "error",
		Message:  "Something went wrong",
		Metadata: model.JSONMap{},
	}
	logRepo.On("List", ctx, teamID, "error", 20, 0).Return([]model.Log{log1}, 1, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, "error", params)

	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "error", resp.Data[0].Level)

	logRepo.AssertExpectations(t)
}

func TestLogService_List_AllLevels(t *testing.T) {
	logRepo := new(tmock.MockLogRepository)
	svc := NewLogService(logRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	log1 := model.Log{ID: uuid.New(), TeamID: teamID, Level: "info", Message: "Info message", Metadata: model.JSONMap{}}
	log2 := model.Log{ID: uuid.New(), TeamID: teamID, Level: "error", Message: "Error message", Metadata: model.JSONMap{}}
	logRepo.On("List", ctx, teamID, "", 20, 0).Return([]model.Log{log1, log2}, 2, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, "", params)

	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Data, 2)

	logRepo.AssertExpectations(t)
}

func TestLogService_List_EmptyResult(t *testing.T) {
	logRepo := new(tmock.MockLogRepository)
	svc := NewLogService(logRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	logRepo.On("List", ctx, teamID, "debug", 20, 0).Return([]model.Log{}, 0, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, "debug", params)

	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Data)
	assert.False(t, resp.HasMore)

	logRepo.AssertExpectations(t)
}

func TestLogService_List_Pagination(t *testing.T) {
	logRepo := new(tmock.MockLogRepository)
	svc := NewLogService(logRepo)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	// Page 2, requesting 10 per page, total of 25 records.
	log1 := model.Log{ID: uuid.New(), TeamID: teamID, Level: "info", Message: "Page 2 log", Metadata: model.JSONMap{}}
	logRepo.On("List", ctx, teamID, "", 10, 10).Return([]model.Log{log1}, 25, nil)

	params := &dto.PaginationParams{Page: 2, PerPage: 10}
	resp, err := svc.List(ctx, teamID, "", params)

	require.NoError(t, err)
	assert.Equal(t, 25, resp.Total)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 3, resp.TotalPages)
	assert.True(t, resp.HasMore)

	logRepo.AssertExpectations(t)
}
