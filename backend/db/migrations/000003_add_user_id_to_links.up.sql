ALTER TABLE links
    ADD COLUMN user_id BIGINT NULL REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX links_user_id_created_at_idx ON links (user_id, created_at DESC);