CREATE TABLE page_views (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  path        TEXT        NOT NULL,
  user_id     UUID        REFERENCES users(id) ON DELETE SET NULL,
  session_id  TEXT        NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_page_views_created_at ON page_views(created_at);
CREATE INDEX idx_page_views_path       ON page_views(path);
CREATE INDEX idx_page_views_session    ON page_views(session_id, created_at);
