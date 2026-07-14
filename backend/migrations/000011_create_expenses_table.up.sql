CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tourdate_id UUID NOT NULL REFERENCES tourdates(id) ON DELETE CASCADE,
    category TEXT NOT NULL CHECK (category IN ('TRAVEL', 'LODGING', 'CREW', 'GEAR', 'OTHER')),
    amount NUMERIC(10, 2) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX expenses_tourdate_id_idx ON expenses (tourdate_id);
