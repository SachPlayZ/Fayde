-- Goals / OKRs: top-level objectives with key results; tasks can link to a goal.

CREATE TYPE goal_status AS ENUM ('on_track','at_risk','off_track','done','archived');

CREATE TABLE goals (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status      goal_status NOT NULL DEFAULT 'on_track',
  target_date DATE,
  parent_id   UUID REFERENCES goals(id) ON DELETE SET NULL,
  position    DOUBLE PRECISION NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_goals_user ON goals(user_id);

CREATE TABLE key_results (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  goal_id     UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  metric_type TEXT NOT NULL DEFAULT 'percent',  -- percent | number | task_completion
  current_val DOUBLE PRECISION NOT NULL DEFAULT 0,
  target_val  DOUBLE PRECISION NOT NULL DEFAULT 100,
  position    DOUBLE PRECISION NOT NULL DEFAULT 0
);
CREATE INDEX idx_key_results_goal ON key_results(goal_id);

ALTER TABLE tasks ADD COLUMN goal_id UUID REFERENCES goals(id) ON DELETE SET NULL;
CREATE INDEX idx_tasks_goal ON tasks(goal_id);
