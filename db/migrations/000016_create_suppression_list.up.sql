CREATE TABLE suppression_list (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    reason VARCHAR(20) NOT NULL CHECK (reason IN ('bounce', 'complaint', 'unsubscribe', 'manual')),
    details TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, email)
);

CREATE INDEX idx_suppression_list_team_id ON suppression_list(team_id);
CREATE INDEX idx_suppression_list_email ON suppression_list(email);
