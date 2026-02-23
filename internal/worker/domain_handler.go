package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
)

// DNS record type constants used in DomainDNSRecord.RecordType.
const (
	RecordTypeSPF        = "SPF"
	RecordTypeDKIM       = "DKIM"
	RecordTypeMX         = "MX"
	RecordTypeDMARC      = "DMARC"
	RecordTypeReturnPath = "RETURN_PATH"
)

// DNS record verification status constants.
const (
	DNSStatusPending  = "pending"
	DNSStatusVerified = "verified"
	DNSStatusFailed   = "failed"
)

// DomainVerifyHandler processes domain:verify tasks by checking each DNS record
// associated with a domain and updating their verification status.
type DomainVerifyHandler struct {
	domainRepo    postgres.DomainRepository
	dnsRecordRepo postgres.DomainDNSRecordRepository
	logger        *slog.Logger
}

// NewDomainVerifyHandler creates a new DomainVerifyHandler.
func NewDomainVerifyHandler(
	domainRepo postgres.DomainRepository,
	dnsRecordRepo postgres.DomainDNSRecordRepository,
	logger *slog.Logger,
) *DomainVerifyHandler {
	return &DomainVerifyHandler{
		domainRepo:    domainRepo,
		dnsRecordRepo: dnsRecordRepo,
		logger:        logger,
	}
}

// ProcessTask handles the domain:verify task.
func (h *DomainVerifyHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p DomainVerifyPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshalling domain:verify payload: %w", err)
	}

	log := h.logger.With("domain_id", p.DomainID, "team_id", p.TeamID)

	// 1. Get the domain.
	domain, err := h.domainRepo.GetByID(ctx, p.DomainID)
	if err != nil {
		return fmt.Errorf("fetching domain %s: %w", p.DomainID, err)
	}

	// 2. Get all DNS records for this domain.
	records, err := h.dnsRecordRepo.ListByDomainID(ctx, p.DomainID)
	if err != nil {
		return fmt.Errorf("listing DNS records for domain %s: %w", p.DomainID, err)
	}

	if len(records) == 0 {
		log.Warn("domain has no DNS records to verify")
		return nil
	}

	// 3. Verify each record.
	now := time.Now().UTC()
	allCriticalVerified := true

	for i := range records {
		record := &records[i]
		verified, verifyErr := h.verifyRecord(domain.Name, record)

		record.LastCheckedAt = &now
		record.UpdatedAt = now

		if verifyErr != nil {
			log.Warn("DNS verification failed",
				"record_type", record.RecordType,
				"dns_type", record.DNSType,
				"name", record.Name,
				"error", verifyErr,
			)
			record.Status = DNSStatusFailed
		} else if verified {
			log.Info("DNS record verified",
				"record_type", record.RecordType,
				"dns_type", record.DNSType,
				"name", record.Name,
			)
			record.Status = DNSStatusVerified
		} else {
			record.Status = DNSStatusFailed
		}

		if err := h.dnsRecordRepo.Update(ctx, record); err != nil {
			log.Error("failed to update DNS record status", "record_id", record.ID, "error", err)
		}

		// Track whether all critical records (SPF, DKIM, MX) are verified.
		if isCriticalRecord(record.RecordType) && record.Status != DNSStatusVerified {
			allCriticalVerified = false
		}
	}

	// 4. Update domain status based on verification results.
	if allCriticalVerified {
		domain.Status = model.DomainStatusVerified
		log.Info("domain fully verified")
	} else {
		domain.Status = model.DomainStatusFailed
		log.Info("domain verification incomplete, some critical records failed")
	}

	domain.UpdatedAt = now
	if err := h.domainRepo.Update(ctx, domain); err != nil {
		return fmt.Errorf("updating domain status: %w", err)
	}

	return nil
}

// verifyRecord performs a DNS lookup to verify a single DNS record.
func (h *DomainVerifyHandler) verifyRecord(domainName string, record *model.DomainDNSRecord) (bool, error) {
	switch record.RecordType {
	case RecordTypeSPF:
		return h.verifySPF(record.Name, record.Value)
	case RecordTypeDKIM:
		return h.verifyDKIM(record.Name, record.Value)
	case RecordTypeDMARC:
		return h.verifyDMARC(record.Name, record.Value)
	case RecordTypeMX:
		return h.verifyMX(record.Name, record.Value, record.Priority)
	case RecordTypeReturnPath:
		return h.verifyCNAME(record.Name, record.Value)
	default:
		return false, fmt.Errorf("unknown record type: %s", record.RecordType)
	}
}

// verifySPF checks that the expected SPF TXT record is published.
func (h *DomainVerifyHandler) verifySPF(name, expectedValue string) (bool, error) {
	txtRecords, err := net.LookupTXT(name)
	if err != nil {
		return false, fmt.Errorf("SPF TXT lookup for %s: %w", name, err)
	}

	for _, txt := range txtRecords {
		if strings.Contains(txt, "v=spf1") && strings.Contains(txt, expectedValue) {
			return true, nil
		}
	}

	return false, nil
}

// verifyDKIM checks that the expected DKIM TXT record is published at the selector._domainkey subdomain.
func (h *DomainVerifyHandler) verifyDKIM(name, expectedValue string) (bool, error) {
	txtRecords, err := net.LookupTXT(name)
	if err != nil {
		return false, fmt.Errorf("DKIM TXT lookup for %s: %w", name, err)
	}

	// DKIM TXT records can be split across multiple strings; join them.
	for _, txt := range txtRecords {
		if strings.Contains(txt, "v=DKIM1") && strings.Contains(txt, expectedValue) {
			return true, nil
		}
	}

	return false, nil
}

// verifyDMARC checks that the expected DMARC TXT record is published.
func (h *DomainVerifyHandler) verifyDMARC(name, expectedValue string) (bool, error) {
	txtRecords, err := net.LookupTXT(name)
	if err != nil {
		return false, fmt.Errorf("DMARC TXT lookup for %s: %w", name, err)
	}

	for _, txt := range txtRecords {
		if strings.Contains(txt, "v=DMARC1") && strings.Contains(txt, expectedValue) {
			return true, nil
		}
	}

	return false, nil
}

// verifyMX checks that the expected MX record is published with the correct priority.
func (h *DomainVerifyHandler) verifyMX(name, expectedHost string, expectedPriority *int) (bool, error) {
	mxRecords, err := net.LookupMX(name)
	if err != nil {
		return false, fmt.Errorf("MX lookup for %s: %w", name, err)
	}

	for _, mx := range mxRecords {
		host := strings.TrimSuffix(mx.Host, ".")
		expectedTrimmed := strings.TrimSuffix(expectedHost, ".")

		if strings.EqualFold(host, expectedTrimmed) {
			if expectedPriority == nil || int(mx.Pref) == *expectedPriority {
				return true, nil
			}
		}
	}

	return false, nil
}

// verifyCNAME checks that a CNAME record points to the expected value.
func (h *DomainVerifyHandler) verifyCNAME(name, expectedValue string) (bool, error) {
	cname, err := net.LookupCNAME(name)
	if err != nil {
		return false, fmt.Errorf("CNAME lookup for %s: %w", name, err)
	}

	cnameClean := strings.TrimSuffix(cname, ".")
	expectedClean := strings.TrimSuffix(expectedValue, ".")

	return strings.EqualFold(cnameClean, expectedClean), nil
}

// isCriticalRecord returns true for record types that must be verified for the domain
// to be considered fully verified.
func isCriticalRecord(recordType string) bool {
	switch recordType {
	case RecordTypeSPF, RecordTypeDKIM, RecordTypeMX:
		return true
	default:
		return false
	}
}

// uuidToString converts a *uuid.UUID to string, returning empty for nil.
func uuidToString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
