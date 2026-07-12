CREATE TABLE artist_representatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    artist_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    representative_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (artist_id, representative_id)
);

CREATE INDEX artist_representatives_representative_id_idx ON artist_representatives(representative_id);
