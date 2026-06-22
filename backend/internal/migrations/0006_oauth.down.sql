DROP TABLE IF EXISTS email_verifications;

ALTER TABLE users
  DROP COLUMN IF EXISTS email_verified,
  DROP COLUMN IF EXISTS provider,
  DROP COLUMN IF EXISTS provider_id;

ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
