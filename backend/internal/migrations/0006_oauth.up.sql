ALTER TABLE users
  ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS provider       TEXT    NOT NULL DEFAULT 'local',
  ADD COLUMN IF NOT EXISTS provider_id    TEXT;

ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- Grandfather existing users as already verified.
UPDATE users SET email_verified = true;

CREATE TABLE email_verifications (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token      TEXT        NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '24 hours',
  used_at    TIMESTAMPTZ
);

CREATE INDEX idx_ev_token   ON email_verifications(token);
CREATE INDEX idx_ev_user_id ON email_verifications(user_id);
