ALTER TABLE tasks DROP COLUMN IF EXISTS project_id;
DROP INDEX IF EXISTS idx_tasks_project_id;

CREATE TABLE IF NOT EXISTS showcase_tokens (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    token_hash   TEXT NOT NULL UNIQUE,
    token_prefix VARCHAR(20) NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_showcase_tokens_user ON showcase_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_showcase_tokens_hash ON showcase_tokens(token_hash);

ALTER INDEX idx_projects_user RENAME TO idx_showcase_entries_user;
ALTER TABLE projects RENAME TO showcase_entries;

CREATE TABLE IF NOT EXISTS projects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    color       TEXT NOT NULL DEFAULT '#6366f1',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE tasks ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
