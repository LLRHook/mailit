CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audience_id UUID NOT NULL REFERENCES audiences(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    unsubscribed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(audience_id, email)
);

CREATE INDEX idx_contacts_audience_id ON contacts(audience_id);
CREATE INDEX idx_contacts_email ON contacts(email);
