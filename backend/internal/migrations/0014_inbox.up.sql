-- Email-to-task: per-user inbox token for inbound email addresses.

ALTER TABLE users ADD COLUMN inbox_token TEXT;
UPDATE users SET inbox_token = left(replace(gen_random_uuid()::text, '-', ''), 12) WHERE inbox_token IS NULL;
ALTER TABLE users ALTER COLUMN inbox_token SET DEFAULT left(replace(gen_random_uuid()::text, '-', ''), 12);
ALTER TABLE users ADD CONSTRAINT users_inbox_token_key UNIQUE (inbox_token);
