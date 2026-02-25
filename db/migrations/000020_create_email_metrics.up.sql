CREATE TABLE email_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    period_start TIMESTAMPTZ NOT NULL,
    period_type VARCHAR(10) NOT NULL CHECK (period_type IN ('hourly', 'daily')),
    sent INT DEFAULT 0,
    delivered INT DEFAULT 0,
    bounced INT DEFAULT 0,
    failed INT DEFAULT 0,
    opened INT DEFAULT 0,
    clicked INT DEFAULT 0,
    complained INT DEFAULT 0,
    UNIQUE (team_id, period_start, period_type)
);

CREATE INDEX idx_email_metrics_team_period ON email_metrics(team_id, period_type, period_start);
