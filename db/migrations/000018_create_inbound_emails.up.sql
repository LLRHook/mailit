CREATE TABLE inbound_emails (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    domain_id UUID REFERENCES domains(id) ON DELETE SET NULL,
    from_address VARCHAR(255) NOT NULL,
    to_addresses TEXT[] NOT NULL,
    cc_addresses TEXT[],
    subject TEXT,
    html_body TEXT,
    text_body TEXT,
    raw_message TEXT,
    headers JSONB DEFAULT '{}',
    attachments JSONB DEFAULT '[]',
    spam_score FLOAT,
    processed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inbound_emails_team_id ON inbound_emails(team_id);
CREATE INDEX idx_inbound_emails_created_at ON inbound_emails(created_at DESC);
