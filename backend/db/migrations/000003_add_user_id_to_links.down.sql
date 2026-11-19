DROP INDEX IF EXISTS links_user_id_created_at_idx;
ALTER TABLE links DROP COLUMN IF EXISTS user_id;