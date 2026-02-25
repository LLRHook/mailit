package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/worker"
)

const (
	// dkimKeyBits is the default RSA key size for DKIM signing.
	dkimKeyBits = 2048
)

// DomainService defines operations for sending domain management.
type DomainService interface {
	Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateDomainRequest) (*dto.DomainResponse, error)
	List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.DomainResponse], error)
	Get(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID) (*dto.DomainResponse, error)
	Update(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID, req *dto.UpdateDomainRequest) (*dto.DomainResponse, error)
	Delete(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID) error
	Verify(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID) (*dto.DomainResponse, error)
}

type domainService struct {
	domainRepo    postgres.DomainRepository
	dnsRecordRepo postgres.DomainDNSRecordRepository
	asynqClient   *asynq.Client
	dkimSelector  string
	encryptionKey string
}

// NewDomainService creates a new DomainService.
func NewDomainService(
	domainRepo postgres.DomainRepository,
	dnsRecordRepo postgres.DomainDNSRecordRepository,
	asynqClient *asynq.Client,
	dkimSelector string,
	encryptionKey string,
) DomainService {
	return &domainService{
		domainRepo:    domainRepo,
		dnsRecordRepo: dnsRecordRepo,
		asynqClient:   asynqClient,
		dkimSelector:  dkimSelector,
		encryptionKey: encryptionKey,
	}
}

func (s *domainService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateDomainRequest) (*dto.DomainResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	// Check for duplicate domain within the team.
	existing, _ := s.domainRepo.GetByTeamAndName(ctx, teamID, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("domain %s already exists for this team", req.Name)
	}

	// Generate DKIM RSA key pair.
	privateKey, err := rsa.GenerateKey(rand.Reader, dkimKeyBits)
	if err != nil {
		return nil, fmt.Errorf("generating DKIM key pair: %w", err)
	}

	// Encode private key to PEM.
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode public key for DNS TXT record.
	pubDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshalling DKIM public key: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})

	// Store private key. In production this would be encrypted with the master key.
	// For now, store as PEM string (encryption can be layered on top).
	privKeyStr := string(privPEM)

	now := time.Now().UTC()
	selector := s.dkimSelector
	if selector == "" {
		selector = "mailit"
	}

	domain := &model.Domain{
		ID:             uuid.New(),
		TeamID:         teamID,
		Name:           req.Name,
		Status:         model.DomainStatusPending,
		DKIMPrivateKey: &privKeyStr,
		DKIMSelector:   selector,
		OpenTracking:   false,
		ClickTracking:  false,
		TLSPolicy:      "opportunistic",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.domainRepo.Create(ctx, domain); err != nil {
		return nil, fmt.Errorf("creating domain: %w", err)
	}

	// Create DNS records for the domain.
	records := s.buildDNSRecords(domain.ID, req.Name, selector, string(pubPEM), now)
	for i := range records {
		if err := s.dnsRecordRepo.Create(ctx, &records[i]); err != nil {
			return nil, fmt.Errorf("creating DNS record: %w", err)
		}
	}

	// Enqueue verification task.
	s.enqueueVerifyTask(domain.ID, teamID)

	return s.buildDomainResponse(domain, records), nil
}

func (s *domainService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.DomainResponse], error) {
	params.Normalize()

	domains, total, err := s.domainRepo.List(ctx, teamID, params.PerPage, params.Offset())
	if err != nil {
		return nil, fmt.Errorf("listing domains: %w", err)
	}

	data := make([]dto.DomainResponse, 0, len(domains))
	for _, d := range domains {
		records, err := s.dnsRecordRepo.ListByDomainID(ctx, d.ID)
		if err != nil {
			return nil, fmt.Errorf("listing DNS records for domain %s: %w", d.ID, err)
		}
		data = append(data, *s.buildDomainResponse(&d, records))
	}

	totalPages := 0
	if params.PerPage > 0 {
		totalPages = (total + params.PerPage - 1) / params.PerPage
	}

	return &dto.PaginatedResponse[dto.DomainResponse]{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
		HasMore:    params.Page < totalPages,
	}, nil
}

func (s *domainService) Get(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID) (*dto.DomainResponse, error) {
	domain, err := s.domainRepo.GetByTeamAndID(ctx, teamID, domainID)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	records, err := s.dnsRecordRepo.ListByDomainID(ctx, domain.ID)
	if err != nil {
		return nil, fmt.Errorf("listing DNS records: %w", err)
	}

	return s.buildDomainResponse(domain, records), nil
}

func (s *domainService) Update(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID, req *dto.UpdateDomainRequest) (*dto.DomainResponse, error) {
	if err := pkg.Validate(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	domain, err := s.domainRepo.GetByTeamAndID(ctx, teamID, domainID)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	if req.OpenTracking != nil {
		domain.OpenTracking = *req.OpenTracking
	}
	if req.ClickTracking != nil {
		domain.ClickTracking = *req.ClickTracking
	}
	if req.TLSPolicy != nil {
		domain.TLSPolicy = *req.TLSPolicy
	}

	domain.UpdatedAt = time.Now().UTC()

	if err := s.domainRepo.Update(ctx, domain); err != nil {
		return nil, fmt.Errorf("updating domain: %w", err)
	}

	records, err := s.dnsRecordRepo.ListByDomainID(ctx, domain.ID)
	if err != nil {
		return nil, fmt.Errorf("listing DNS records: %w", err)
	}

	return s.buildDomainResponse(domain, records), nil
}

func (s *domainService) Delete(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID) error {
	_, err := s.domainRepo.GetByTeamAndID(ctx, teamID, domainID)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}

	if err := s.dnsRecordRepo.DeleteByDomainID(ctx, domainID); err != nil {
		return fmt.Errorf("deleting DNS records: %w", err)
	}

	if err := s.domainRepo.Delete(ctx, domainID); err != nil {
		return fmt.Errorf("deleting domain: %w", err)
	}

	return nil
}

func (s *domainService) Verify(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID) (*dto.DomainResponse, error) {
	domain, err := s.domainRepo.GetByTeamAndID(ctx, teamID, domainID)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Enqueue a verification task.
	s.enqueueVerifyTask(domain.ID, teamID)

	records, err := s.dnsRecordRepo.ListByDomainID(ctx, domain.ID)
	if err != nil {
		return nil, fmt.Errorf("listing DNS records: %w", err)
	}

	return s.buildDomainResponse(domain, records), nil
}

// buildDNSRecords creates the set of required DNS records for a new domain.
func (s *domainService) buildDNSRecords(domainID uuid.UUID, domainName, selector, pubKeyPEM string, now time.Time) []model.DomainDNSRecord {
	mxPriority := 10

	return []model.DomainDNSRecord{
		{
			ID:         uuid.New(),
			DomainID:   domainID,
			RecordType: "SPF",
			DNSType:    "TXT",
			Name:       domainName,
			Value:      "v=spf1 include:_spf." + domainName + " ~all",
			Status:     model.DomainStatusPending,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			DomainID:   domainID,
			RecordType: "DKIM",
			DNSType:    "TXT",
			Name:       selector + "._domainkey." + domainName,
			Value:      "v=DKIM1; k=rsa; p=" + extractBase64Key(pubKeyPEM),
			Status:     model.DomainStatusPending,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			DomainID:   domainID,
			RecordType: "MX",
			DNSType:    "MX",
			Name:       domainName,
			Value:      "feedback-smtp." + domainName,
			Priority:   &mxPriority,
			Status:     model.DomainStatusPending,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			DomainID:   domainID,
			RecordType: "DMARC",
			DNSType:    "TXT",
			Name:       "_dmarc." + domainName,
			Value:      "v=DMARC1; p=none;",
			Status:     model.DomainStatusPending,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			DomainID:   domainID,
			RecordType: "RETURN_PATH",
			DNSType:    "CNAME",
			Name:       "bounce." + domainName,
			Value:      "feedback-smtp." + domainName,
			Status:     model.DomainStatusPending,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
}

// enqueueVerifyTask creates an asynq task to verify the domain's DNS records.
func (s *domainService) enqueueVerifyTask(domainID, teamID uuid.UUID) {
	payload, _ := json.Marshal(map[string]string{
		"domain_id": domainID.String(),
		"team_id":   teamID.String(),
	})
	task := asynq.NewTask(worker.TaskDomainVerify, payload)
	_, _ = s.asynqClient.Enqueue(task, asynq.Queue(worker.QueueDefault), asynq.MaxRetry(3))
}

// buildDomainResponse converts a domain model and its DNS records to a DTO response.
func (s *domainService) buildDomainResponse(domain *model.Domain, records []model.DomainDNSRecord) *dto.DomainResponse {
	dnsRecords := make([]dto.DNSRecordResponse, 0, len(records))
	for _, r := range records {
		dnsRecords = append(dnsRecords, dto.DNSRecordResponse{
			Type:     r.RecordType,
			Name:     r.Name,
			Value:    r.Value,
			Priority: r.Priority,
			Status:   r.Status,
			TTL:      "Auto",
		})
	}

	return &dto.DomainResponse{
		ID:        domain.ID.String(),
		Name:      domain.Name,
		Status:    domain.Status,
		Region:    domain.Region,
		Records:   dnsRecords,
		CreatedAt: domain.CreatedAt.Format(time.RFC3339),
	}
}

// extractBase64Key strips the PEM headers and newlines to produce a raw base64 key string.
func extractBase64Key(pemStr string) string {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return ""
	}
	// Encode the raw DER bytes as a continuous base64 string.
	// The pem package already provides the base64 content in block.Bytes,
	// but we need to re-encode to remove line breaks.
	import64 := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: block.Bytes})
	// Strip header, footer, and newlines.
	result := string(import64)
	result = result[len("-----BEGIN PUBLIC KEY-----\n"):]
	result = result[:len(result)-len("-----END PUBLIC KEY-----\n")]
	// Remove remaining newlines.
	cleaned := ""
	for _, c := range result {
		if c != '\n' && c != '\r' {
			cleaned += string(c)
		}
	}
	return cleaned
}
