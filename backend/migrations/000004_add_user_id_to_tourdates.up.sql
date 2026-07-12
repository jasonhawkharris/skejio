ALTER TABLE tourdates ADD COLUMN user_id UUID NOT NULL REFERENCES users(id);
CREATE INDEX tourdates_user_id_idx ON tourdates(user_id);
