CREATE TABLE tourdates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL,
    city TEXT NOT NULL,
    state TEXT,
    venue TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
