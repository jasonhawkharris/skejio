CREATE TABLE financials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tourdate_id UUID NOT NULL UNIQUE REFERENCES tourdates(id) ON DELETE CASCADE,
    fee NUMERIC(10, 2),
    tips NUMERIC(10, 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
