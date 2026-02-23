CREATE TABLE email_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_id UUID NOT NULL REFERENCES emails(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('queued', 'sent', 'delivered', 'bounced', 'opened', 'clicked', 'complained', 'unsubscribed', 'failed')),
    payload JSONB DEFAULT '{}',
    recipient VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_events_email_id ON email_events(email_id);
CREATE INDEX idx_email_events_type ON email_events(type);
CREATE INDEX idx_email_events_created_at ON email_events(created_at DESC);
