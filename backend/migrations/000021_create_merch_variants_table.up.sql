CREATE TABLE merch_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merch_id UUID NOT NULL REFERENCES merch(id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    inventory_count INTEGER NOT NULL DEFAULT 0 CHECK (inventory_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (merch_id, label)
);

CREATE INDEX merch_variants_merch_id_idx ON merch_variants (merch_id);
