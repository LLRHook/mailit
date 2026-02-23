CREATE TABLE domain_dns_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    record_type VARCHAR(10) NOT NULL CHECK (record_type IN ('SPF', 'DKIM', 'MX', 'DMARC', 'RETURN_PATH')),
    dns_type VARCHAR(10) NOT NULL CHECK (dns_type IN ('TXT', 'MX', 'CNAME', 'A', 'AAAA')),
    name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    priority INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'failed')),
    last_checked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_domain_dns_records_domain_id ON domain_dns_records(domain_id);
