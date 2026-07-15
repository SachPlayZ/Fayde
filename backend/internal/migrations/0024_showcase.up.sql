ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(50);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username IS NOT NULL;

CREATE TABLE IF NOT EXISTS showcase_entries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title         TEXT NOT NULL,
    tagline       TEXT NOT NULL DEFAULT '',
    problem       TEXT NOT NULL DEFAULT '',
    solution      TEXT NOT NULL DEFAULT '',
    tech_stack    JSONB NOT NULL DEFAULT '[]',
    demo_url      TEXT,
    repo_url      TEXT,
    live_url      TEXT,
    logo_s3_key   TEXT,
    banner_s3_key TEXT,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_showcase_entries_user ON showcase_entries(user_id);

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
