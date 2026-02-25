package service

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/testutil"
	tmock "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func newDomainTestDeps(t *testing.T) (*tmock.MockDomainRepository, *tmock.MockDomainDNSRecordRepository, *asynq.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: mr.Addr()})
	return new(tmock.MockDomainRepository), new(tmock.MockDomainDNSRecordRepository), asynqClient
}

func TestDomainService_Create_HappyPath(t *testing.T) {
	domainRepo, dnsRepo, asynqClient := newDomainTestDeps(t)
	svc := NewDomainService(domainRepo, dnsRepo, asynqClient, "mailit", "")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	// No duplicate domain.
	domainRepo.On("GetByTeamAndName", ctx, teamID, "example.com").Return(nil, postgres.ErrNotFound)
	domainRepo.On("Create", ctx, mock.AnythingOfType("*model.Domain")).Return(nil)
	// 5 DNS records are created (SPF, DKIM, MX, DMARC, RETURN_PATH).
	dnsRepo.On("Create", ctx, mock.AnythingOfType("*model.DomainDNSRecord")).Return(nil).Times(5)

	req := &dto.CreateDomainRequest{Name: "example.com"}
	resp, err := svc.Create(ctx, teamID, req)

	require.NoError(t, err)
	assert.Equal(t, "example.com", resp.Name)
	assert.Equal(t, model.DomainStatusPending, resp.Status)
	assert.Len(t, resp.Records, 5)

	domainRepo.AssertExpectations(t)
	dnsRepo.AssertExpectations(t)
}

func TestDomainService_Create_DuplicateDomain(t *testing.T) {
	domainRepo, dnsRepo, asynqClient := newDomainTestDeps(t)
	svc := NewDomainService(domainRepo, dnsRepo, asynqClient, "mailit", "")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	existing := testutil.NewTestDomain()
	domainRepo.On("GetByTeamAndName", ctx, teamID, "example.com").Return(existing, nil)

	req := &dto.CreateDomainRequest{Name: "example.com"}
	resp, err := svc.Create(ctx, teamID, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	domainRepo.AssertExpectations(t)
}

func TestDomainService_List_Paginated(t *testing.T) {
	domainRepo, dnsRepo, asynqClient := newDomainTestDeps(t)
	svc := NewDomainService(domainRepo, dnsRepo, asynqClient, "mailit", "")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	domain1 := *testutil.NewTestDomain()
	domainRepo.On("List", ctx, teamID, 20, 0).Return([]model.Domain{domain1}, 1, nil)
	dnsRepo.On("ListByDomainID", ctx, domain1.ID).Return([]model.DomainDNSRecord{}, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, params)

	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Data, 1)

	domainRepo.AssertExpectations(t)
	dnsRepo.AssertExpectations(t)
}

func TestDomainService_Get_HappyPath(t *testing.T) {
	domainRepo, dnsRepo, asynqClient := newDomainTestDeps(t)
	svc := NewDomainService(domainRepo, dnsRepo, asynqClient, "mailit", "")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	domain := testutil.NewTestDomain()
	domainRepo.On("GetByTeamAndID", ctx, teamID, domain.ID).Return(domain, nil)
	dnsRepo.On("ListByDomainID", ctx, domain.ID).Return([]model.DomainDNSRecord{}, nil)

	resp, err := svc.Get(ctx, teamID, domain.ID)

	require.NoError(t, err)
	assert.Equal(t, domain.ID.String(), resp.ID)
	assert.Equal(t, "example.com", resp.Name)

	domainRepo.AssertExpectations(t)
	dnsRepo.AssertExpectations(t)
}

func TestDomainService_Update_TrackingSettings(t *testing.T) {
	domainRepo, dnsRepo, asynqClient := newDomainTestDeps(t)
	svc := NewDomainService(domainRepo, dnsRepo, asynqClient, "mailit", "")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	domain := testutil.NewTestDomain()
	domainRepo.On("GetByTeamAndID", ctx, teamID, domain.ID).Return(domain, nil)
	domainRepo.On("Update", ctx, mock.AnythingOfType("*model.Domain")).Return(nil)
	dnsRepo.On("ListByDomainID", ctx, domain.ID).Return([]model.DomainDNSRecord{}, nil)

	req := &dto.UpdateDomainRequest{
		OpenTracking:  testutil.BoolPtr(true),
		ClickTracking: testutil.BoolPtr(true),
	}
	resp, err := svc.Update(ctx, teamID, domain.ID, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)

	domainRepo.AssertExpectations(t)
	dnsRepo.AssertExpectations(t)
}

func TestDomainService_Delete_HappyPath(t *testing.T) {
	domainRepo, dnsRepo, asynqClient := newDomainTestDeps(t)
	svc := NewDomainService(domainRepo, dnsRepo, asynqClient, "mailit", "")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	domain := testutil.NewTestDomain()
	domainRepo.On("GetByTeamAndID", ctx, teamID, domain.ID).Return(domain, nil)
	dnsRepo.On("DeleteByDomainID", ctx, domain.ID).Return(nil)
	domainRepo.On("Delete", ctx, domain.ID).Return(nil)

	err := svc.Delete(ctx, teamID, domain.ID)

	require.NoError(t, err)

	domainRepo.AssertExpectations(t)
	dnsRepo.AssertExpectations(t)
}

func TestDomainService_Verify_EnqueuesTask(t *testing.T) {
	domainRepo, dnsRepo, asynqClient := newDomainTestDeps(t)
	svc := NewDomainService(domainRepo, dnsRepo, asynqClient, "mailit", "")
	ctx := context.Background()
	teamID := testutil.TestTeamID

	domain := testutil.NewTestDomain()
	domainRepo.On("GetByTeamAndID", ctx, teamID, domain.ID).Return(domain, nil)
	dnsRepo.On("ListByDomainID", ctx, domain.ID).Return([]model.DomainDNSRecord{}, nil)

	resp, err := svc.Verify(ctx, teamID, domain.ID)

	require.NoError(t, err)
	assert.Equal(t, domain.ID.String(), resp.ID)

	domainRepo.AssertExpectations(t)
	dnsRepo.AssertExpectations(t)
}

func TestDomainService_Get_NotFound(t *testing.T) {
	domainRepo, dnsRepo, asynqClient := newDomainTestDeps(t)
	svc := NewDomainService(domainRepo, dnsRepo, asynqClient, "mailit", "")
	ctx := context.Background()
	teamID := testutil.TestTeamID
	badID := uuid.New()

	domainRepo.On("GetByTeamAndID", ctx, teamID, badID).Return(nil, postgres.ErrNotFound)

	resp, err := svc.Get(ctx, teamID, badID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	domainRepo.AssertExpectations(t)
}
