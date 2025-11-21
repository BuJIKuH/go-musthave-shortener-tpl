DROP INDEX IF EXISTS idx_urls_user_id;
DROP INDEX IF EXISTS idx_urls_is_deleted;

ALTER TABLE urls
    DROP COLUMN IF EXISTS user_id,
    DROP COLUMN IF EXISTS is_deleted,
    DROP COLUMN IF EXISTS created_at;
