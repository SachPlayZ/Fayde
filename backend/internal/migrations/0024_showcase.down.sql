DROP TABLE IF EXISTS showcase_tokens;
DROP TABLE IF EXISTS showcase_entries;
DROP INDEX IF EXISTS idx_users_username;
ALTER TABLE users DROP COLUMN IF EXISTS username;
