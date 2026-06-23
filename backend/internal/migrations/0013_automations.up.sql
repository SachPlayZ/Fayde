-- Automation rules: if-this-then-that on task events.

CREATE TABLE automations (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name       TEXT NOT NULL,
  enabled    BOOLEAN NOT NULL DEFAULT true,
  trigger    JSONB NOT NULL,                 -- {"event":"status_changed","to":"done"}
  conditions JSONB NOT NULL DEFAULT '[]',    -- [{"field":"priority","op":"eq","value":"high"}]
  actions    JSONB NOT NULL DEFAULT '[]',    -- [{"type":"set_priority","value":"low"}]
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_automations_user ON automations(user_id);
