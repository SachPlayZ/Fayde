-- Custom per-task reminders + notification snooze.

CREATE TABLE reminders (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  task_id    UUID REFERENCES tasks(id) ON DELETE CASCADE,
  remind_at  TIMESTAMPTZ NOT NULL,
  note       TEXT NOT NULL DEFAULT '',
  sent       BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_reminders_due ON reminders(remind_at) WHERE NOT sent;

ALTER TABLE notifications ADD COLUMN snoozed_until TIMESTAMPTZ;
