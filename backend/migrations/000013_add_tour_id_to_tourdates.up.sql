ALTER TABLE tourdates
    ADD COLUMN tour_id UUID REFERENCES tours(id) ON DELETE SET NULL;

CREATE INDEX tourdates_tour_id_idx ON tourdates (tour_id);
