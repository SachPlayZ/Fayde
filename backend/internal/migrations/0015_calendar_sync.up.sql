-- Create calendar_connections table to hold user google credentials
CREATE TABLE IF NOT EXISTS calendar_connections (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider      TEXT NOT NULL DEFAULT 'google',
  email         TEXT NOT NULL,
  access_token  TEXT NOT NULL,
  refresh_token TEXT NOT NULL,
  expiry        TIMESTAMPTZ NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, provider)
);

-- Add external_event_id to tasks to keep track of google calendar events
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS external_event_id TEXT;
