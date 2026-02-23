CREATE TABLE contact_properties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    label VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'string' CHECK (type IN ('string', 'number', 'boolean', 'date')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, name)
);

CREATE TABLE contact_property_values (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES contact_properties(id) ON DELETE CASCADE,
    value TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(contact_id, property_id)
);

CREATE INDEX idx_contact_property_values_contact_id ON contact_property_values(contact_id);
