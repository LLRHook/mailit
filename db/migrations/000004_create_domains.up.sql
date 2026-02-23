CREATE TABLE domains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'failed')),
    region VARCHAR(50),
    dkim_private_key TEXT,
    dkim_selector VARCHAR(63) NOT NULL DEFAULT 'mailit',
    open_tracking BOOLEAN NOT NULL DEFAULT true,
    click_tracking BOOLEAN NOT NULL DEFAULT true,
    tls_policy VARCHAR(20) NOT NULL DEFAULT 'opportunistic' CHECK (tls_policy IN ('opportunistic', 'enforce')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, name)
);

CREATE INDEX idx_domains_team_id ON domains(team_id);
CREATE INDEX idx_domains_name ON domains(name);

-- Add the FK from api_keys.domain_id to domains now that the table exists
ALTER TABLE api_keys ADD CONSTRAINT fk_api_keys_domain_id FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE SET NULL;
