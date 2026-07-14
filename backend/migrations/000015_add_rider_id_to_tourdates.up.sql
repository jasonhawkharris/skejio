ALTER TABLE tourdates
    ADD COLUMN rider_id UUID REFERENCES riders(id) ON DELETE SET NULL;

CREATE INDEX tourdates_rider_id_idx ON tourdates (rider_id);
