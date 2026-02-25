-- Add compound index on domain_dns_records for faster lookup by domain + record type.
CREATE INDEX IF NOT EXISTS idx_domain_dns_records_domain_type
    ON domain_dns_records (domain_id, record_type);

-- Add index on email_events.recipient for webhook/tracking lookups.
CREATE INDEX IF NOT EXISTS idx_email_events_recipient
    ON email_events (recipient);

-- Add index on contacts for audience + email uniqueness lookups.
CREATE UNIQUE INDEX IF NOT EXISTS idx_contacts_audience_email
    ON contacts (audience_id, email);

-- Add index on broadcasts for scheduled_at to support time-based queries.
CREATE INDEX IF NOT EXISTS idx_broadcasts_scheduled_at
    ON broadcasts (scheduled_at) WHERE scheduled_at IS NOT NULL;
