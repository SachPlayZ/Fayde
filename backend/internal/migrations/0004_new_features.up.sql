-- effort_points on tasks
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS effort_points INT;

-- projects
CREATE TABLE IF NOT EXISTS projects (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  color       TEXT NOT NULL DEFAULT '#6366f1',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL;

-- time_entries
CREATE TABLE IF NOT EXISTS time_entries (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id          UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  started_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at         TIMESTAMPTZ,
  duration_seconds INT,
  note             TEXT NOT NULL DEFAULT '',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_time_entries_task ON time_entries(task_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_user ON time_entries(user_id);

-- sprints
CREATE TABLE IF NOT EXISTS sprints (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name       TEXT NOT NULL,
  start_date DATE NOT NULL,
  end_date   DATE NOT NULL,
  goal       TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS sprint_tasks (
  sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
  task_id   UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  PRIMARY KEY (sprint_id, task_id)
);

-- task_templates
CREATE TABLE IF NOT EXISTS task_templates (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  title         TEXT NOT NULL,
  description   TEXT NOT NULL DEFAULT '',
  status        TEXT NOT NULL DEFAULT 'todo',
  priority      TEXT NOT NULL DEFAULT 'medium',
  effort_points INT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- custom_fields
CREATE TABLE IF NOT EXISTS custom_field_definitions (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name       TEXT NOT NULL,
  field_type TEXT NOT NULL,
  options    JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, name)
);
CREATE TABLE IF NOT EXISTS custom_field_values (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id  UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  field_id UUID NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
  value    TEXT NOT NULL DEFAULT '',
  UNIQUE(task_id, field_id)
);

-- watchers
CREATE TABLE IF NOT EXISTS task_watchers (
  task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  PRIMARY KEY (task_id, user_id)
);

-- saved filters
CREATE TABLE IF NOT EXISTS saved_filters (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name       TEXT NOT NULL,
  params     JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- api_tokens
CREATE TABLE IF NOT EXISTS api_tokens (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  token_hash   TEXT NOT NULL UNIQUE,
  last_used_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- totp
CREATE TABLE IF NOT EXISTS totp_secrets (
  user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  secret     TEXT NOT NULL,
  enabled    BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN NOT NULL DEFAULT false;

-- github links
CREATE TABLE IF NOT EXISTS github_task_links (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id      UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  repo         TEXT NOT NULL,
  issue_number INT,
  pr_number    INT,
  issue_url    TEXT,
  pr_url       TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- outbound webhooks
CREATE TABLE IF NOT EXISTS outbound_webhooks (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name       TEXT NOT NULL,
  url        TEXT NOT NULL,
  events     TEXT[] NOT NULL DEFAULT '{}',
  secret     TEXT NOT NULL DEFAULT '',
  enabled    BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- public share tokens
CREATE TABLE IF NOT EXISTS task_share_tokens (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id    UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  token      TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- pomodoro sessions
CREATE TABLE IF NOT EXISTS pomodoro_sessions (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id          UUID REFERENCES tasks(id) ON DELETE SET NULL,
  user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  duration_minutes INT NOT NULL DEFAULT 25,
  completed        BOOLEAN NOT NULL DEFAULT false,
  started_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at         TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pomodoro_user ON pomodoro_sessions(user_id);
