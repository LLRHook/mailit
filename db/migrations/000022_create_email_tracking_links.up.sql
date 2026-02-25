CREATE TABLE email_tracking_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_id UUID NOT NULL REFERENCES emails(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id),
    type VARCHAR(20) NOT NULL,
    original_url TEXT,
    recipient TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_tracking_links_email_id ON email_tracking_links(email_id);
CREATE INDEX idx_email_tracking_links_type ON email_tracking_links(type);
