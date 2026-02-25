CREATE TABLE contact_import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    audience_id UUID NOT NULL REFERENCES audiences(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_rows INT NOT NULL DEFAULT 0,
    processed_rows INT NOT NULL DEFAULT 0,
    created_rows INT NOT NULL DEFAULT 0,
    updated_rows INT NOT NULL DEFAULT 0,
    skipped_rows INT NOT NULL DEFAULT 0,
    failed_rows INT NOT NULL DEFAULT 0,
    error TEXT,
    csv_data TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contact_import_jobs_team_id ON contact_import_jobs(team_id);
CREATE INDEX idx_contact_import_jobs_audience_id ON contact_import_jobs(audience_id);
