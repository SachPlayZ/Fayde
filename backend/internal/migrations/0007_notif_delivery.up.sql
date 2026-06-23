-- Step 0: unified notification delivery — per-channel prefs + web push subscriptions + chat target.

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS notif_prefs     JSONB NOT NULL DEFAULT '{"in_app":true,"email":false,"web_push":true,"chat":false}',
  ADD COLUMN IF NOT EXISTS notif_chat_url  TEXT,
  ADD COLUMN IF NOT EXISTS notif_chat_kind TEXT; -- 'slack' | 'discord'

CREATE TABLE IF NOT EXISTS push_subscriptions (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  endpoint   TEXT NOT NULL,
  p256dh     TEXT NOT NULL,
  auth       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, endpoint)
);

CREATE INDEX IF NOT EXISTS idx_push_subs_user ON push_subscriptions(user_id);
