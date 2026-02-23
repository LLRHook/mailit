CREATE TABLE emails (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    domain_id UUID REFERENCES domains(id) ON DELETE SET NULL,
    from_address VARCHAR(255) NOT NULL,
    to_addresses TEXT[] NOT NULL,
    cc_addresses TEXT[],
    bcc_addresses TEXT[],
    reply_to VARCHAR(255),
    subject TEXT NOT NULL,
    html_body TEXT,
    text_body TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'scheduled', 'sending', 'sent', 'delivered', 'bounced', 'failed', 'cancelled')),
    scheduled_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    tags TEXT[],
    headers JSONB DEFAULT '{}',
    attachments JSONB DEFAULT '[]',
    idempotency_key VARCHAR(255),
    message_id VARCHAR(255),
    last_error TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, idempotency_key)
);

CREATE INDEX idx_emails_team_id ON emails(team_id);
CREATE INDEX idx_emails_status ON emails(status);
CREATE INDEX idx_emails_from_address ON emails(from_address);
CREATE INDEX idx_emails_created_at ON emails(created_at DESC);
CREATE INDEX idx_emails_message_id ON emails(message_id);
CREATE INDEX idx_emails_scheduled_at ON emails(scheduled_at) WHERE scheduled_at IS NOT NULL;
