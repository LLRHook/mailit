CREATE TABLE broadcasts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    audience_id UUID REFERENCES audiences(id) ON DELETE SET NULL,
    segment_id UUID REFERENCES segments(id) ON DELETE SET NULL,
    template_id UUID REFERENCES templates(id) ON DELETE SET NULL,
    from_address VARCHAR(255),
    subject TEXT,
    html_body TEXT,
    text_body TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'queued', 'sending', 'sent', 'cancelled')),
    scheduled_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    total_recipients INTEGER DEFAULT 0,
    sent_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_broadcasts_team_id ON broadcasts(team_id);
CREATE INDEX idx_broadcasts_status ON broadcasts(status);
