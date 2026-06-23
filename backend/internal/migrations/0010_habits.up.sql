-- Habits + streaks: recurring habits distinct from tasks, with per-day completion logs.

CREATE TABLE habits (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name              TEXT NOT NULL,
  cadence           TEXT NOT NULL DEFAULT 'daily',  -- daily | weekly
  target_per_period INT  NOT NULL DEFAULT 1,
  color             TEXT,
  position          DOUBLE PRECISION NOT NULL DEFAULT 0,
  archived          BOOLEAN NOT NULL DEFAULT false,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_habits_user ON habits(user_id);

CREATE TABLE habit_logs (
  habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
  log_date DATE NOT NULL,
  count    INT  NOT NULL DEFAULT 1,
  PRIMARY KEY (habit_id, log_date)
);
