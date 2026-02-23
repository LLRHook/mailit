CREATE TABLE segments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audience_id UUID NOT NULL REFERENCES audiences(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    conditions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE segment_contacts (
    segment_id UUID NOT NULL REFERENCES segments(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (segment_id, contact_id)
);

CREATE INDEX idx_segments_audience_id ON segments(audience_id);
CREATE INDEX idx_segment_contacts_contact_id ON segment_contacts(contact_id);
