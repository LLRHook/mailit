package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailit-dev/mailit/internal/model"
)

type domainRepository struct {
	pool *pgxpool.Pool
}

// NewDomainRepository creates a new DomainRepository backed by PostgreSQL.
func NewDomainRepository(pool *pgxpool.Pool) DomainRepository {
	return &domainRepository{pool: pool}
}

const domainColumns = `id, team_id, name, status, region, dkim_private_key, dkim_selector, open_tracking, click_tracking, tls_policy, created_at, updated_at`

func scanDomain(row pgx.Row) (*model.Domain, error) {
	d := &model.Domain{}
	err := row.Scan(
		&d.ID, &d.TeamID, &d.Name, &d.Status, &d.Region,
		&d.DKIMPrivateKey, &d.DKIMSelector, &d.OpenTracking, &d.ClickTracking,
		&d.TLSPolicy, &d.CreatedAt, &d.UpdatedAt,
	)
	return d, err
}

func (r *domainRepository) Create(ctx context.Context, domain *model.Domain) error {
	query := fmt.Sprintf(`
		INSERT INTO domains (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING %s`, domainColumns, domainColumns)

	row := r.pool.QueryRow(ctx, query,
		domain.ID, domain.TeamID, domain.Name, domain.Status, domain.Region,
		domain.DKIMPrivateKey, domain.DKIMSelector, domain.OpenTracking, domain.ClickTracking,
		domain.TLSPolicy, domain.CreatedAt, domain.UpdatedAt,
	)
	scanned, err := scanDomain(row)
	if err != nil {
		return fmt.Errorf("create domain: %w", err)
	}
	*domain = *scanned
	return nil
}

func (r *domainRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Domain, error) {
	query := fmt.Sprintf(`SELECT %s FROM domains WHERE id = $1`, domainColumns)

	d, err := scanDomain(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("domain")
		}
		return nil, fmt.Errorf("get domain by id: %w", err)
	}
	return d, nil
}

func (r *domainRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Domain, error) {
	query := fmt.Sprintf(`SELECT %s FROM domains WHERE team_id = $1 AND id = $2`, domainColumns)

	d, err := scanDomain(r.pool.QueryRow(ctx, query, teamID, id))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("domain")
		}
		return nil, fmt.Errorf("get domain by team and id: %w", err)
	}
	return d, nil
}

func (r *domainRepository) GetByTeamAndName(ctx context.Context, teamID uuid.UUID, name string) (*model.Domain, error) {
	query := fmt.Sprintf(`SELECT %s FROM domains WHERE team_id = $1 AND name = $2`, domainColumns)

	d, err := scanDomain(r.pool.QueryRow(ctx, query, teamID, name))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("domain")
		}
		return nil, fmt.Errorf("get domain by team and name: %w", err)
	}
	return d, nil
}

func (r *domainRepository) GetVerifiedByName(ctx context.Context, name string) (*model.Domain, error) {
	query := fmt.Sprintf(`SELECT %s FROM domains WHERE name = $1 AND status = 'verified' LIMIT 1`, domainColumns)

	d, err := scanDomain(r.pool.QueryRow(ctx, query, name))
	if err != nil {
		if isNoRows(err) {
			return nil, notFound("domain")
		}
		return nil, fmt.Errorf("get verified domain by name: %w", err)
	}
	return d, nil
}

func (r *domainRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Domain, int, error) {
	countQuery := `SELECT COUNT(*) FROM domains WHERE team_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, teamID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count domains: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT %s FROM domains WHERE team_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, domainColumns)

	rows, err := r.pool.Query(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list domains: %w", err)
	}
	defer rows.Close()

	domains, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Domain, error) {
		var d model.Domain
		err := row.Scan(
			&d.ID, &d.TeamID, &d.Name, &d.Status, &d.Region,
			&d.DKIMPrivateKey, &d.DKIMSelector, &d.OpenTracking, &d.ClickTracking,
			&d.TLSPolicy, &d.CreatedAt, &d.UpdatedAt,
		)
		return d, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collect domains: %w", err)
	}

	return domains, total, nil
}

func (r *domainRepository) Update(ctx context.Context, domain *model.Domain) error {
	query := fmt.Sprintf(`
		UPDATE domains
		SET name = $2, status = $3, region = $4, dkim_private_key = $5, dkim_selector = $6,
		    open_tracking = $7, click_tracking = $8, tls_policy = $9, updated_at = $10
		WHERE id = $1
		RETURNING %s`, domainColumns)

	row := r.pool.QueryRow(ctx, query,
		domain.ID, domain.Name, domain.Status, domain.Region,
		domain.DKIMPrivateKey, domain.DKIMSelector, domain.OpenTracking, domain.ClickTracking,
		domain.TLSPolicy, domain.UpdatedAt,
	)
	scanned, err := scanDomain(row)
	if err != nil {
		if isNoRows(err) {
			return notFound("domain")
		}
		return fmt.Errorf("update domain: %w", err)
	}
	*domain = *scanned
	return nil
}

func (r *domainRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM domains WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete domain: %w", err)
	}
	if result.RowsAffected() == 0 {
		return notFound("domain")
	}
	return nil
}

// --- DomainDNSRecordRepository ---

type domainDNSRecordRepository struct {
	pool *pgxpool.Pool
}

// NewDomainDNSRecordRepository creates a new DomainDNSRecordRepository backed by PostgreSQL.
func NewDomainDNSRecordRepository(pool *pgxpool.Pool) DomainDNSRecordRepository {
	return &domainDNSRecordRepository{pool: pool}
}

const dnsRecordColumns = `id, domain_id, record_type, dns_type, name, value, priority, status, last_checked_at, created_at, updated_at`

func (r *domainDNSRecordRepository) Create(ctx context.Context, record *model.DomainDNSRecord) error {
	query := fmt.Sprintf(`
		INSERT INTO domain_dns_records (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING %s`, dnsRecordColumns, dnsRecordColumns)

	return r.pool.QueryRow(ctx, query,
		record.ID, record.DomainID, record.RecordType, record.DNSType, record.Name, record.Value,
		record.Priority, record.Status, record.LastCheckedAt, record.CreatedAt, record.UpdatedAt,
	).Scan(
		&record.ID, &record.DomainID, &record.RecordType, &record.DNSType, &record.Name, &record.Value,
		&record.Priority, &record.Status, &record.LastCheckedAt, &record.CreatedAt, &record.UpdatedAt,
	)
}

func (r *domainDNSRecordRepository) ListByDomainID(ctx context.Context, domainID uuid.UUID) ([]model.DomainDNSRecord, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM domain_dns_records WHERE domain_id = $1
		ORDER BY created_at ASC`, dnsRecordColumns)

	rows, err := r.pool.Query(ctx, query, domainID)
	if err != nil {
		return nil, fmt.Errorf("list domain dns records: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.DomainDNSRecord, error) {
		var rec model.DomainDNSRecord
		err := row.Scan(
			&rec.ID, &rec.DomainID, &rec.RecordType, &rec.DNSType, &rec.Name, &rec.Value,
			&rec.Priority, &rec.Status, &rec.LastCheckedAt, &rec.CreatedAt, &rec.UpdatedAt,
		)
		return rec, err
	})
}

func (r *domainDNSRecordRepository) Update(ctx context.Context, record *model.DomainDNSRecord) error {
	query := fmt.Sprintf(`
		UPDATE domain_dns_records
		SET record_type = $2, dns_type = $3, name = $4, value = $5, priority = $6,
		    status = $7, last_checked_at = $8, updated_at = $9
		WHERE id = $1
		RETURNING %s`, dnsRecordColumns)

	err := r.pool.QueryRow(ctx, query,
		record.ID, record.RecordType, record.DNSType, record.Name, record.Value,
		record.Priority, record.Status, record.LastCheckedAt, record.UpdatedAt,
	).Scan(
		&record.ID, &record.DomainID, &record.RecordType, &record.DNSType, &record.Name, &record.Value,
		&record.Priority, &record.Status, &record.LastCheckedAt, &record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return notFound("domain dns record")
		}
		return fmt.Errorf("update domain dns record: %w", err)
	}
	return nil
}

func (r *domainDNSRecordRepository) DeleteByDomainID(ctx context.Context, domainID uuid.UUID) error {
	query := `DELETE FROM domain_dns_records WHERE domain_id = $1`

	_, err := r.pool.Exec(ctx, query, domainID)
	if err != nil {
		return fmt.Errorf("delete domain dns records by domain: %w", err)
	}
	return nil
}
