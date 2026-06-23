ALTER TABLE users DROP CONSTRAINT IF EXISTS users_inbox_token_key;
ALTER TABLE users DROP COLUMN IF EXISTS inbox_token;
