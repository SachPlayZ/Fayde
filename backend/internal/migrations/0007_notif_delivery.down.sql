DROP TABLE IF EXISTS push_subscriptions;

ALTER TABLE users
  DROP COLUMN IF EXISTS notif_prefs,
  DROP COLUMN IF EXISTS notif_chat_url,
  DROP COLUMN IF EXISTS notif_chat_kind;
