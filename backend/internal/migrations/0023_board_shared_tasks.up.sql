CREATE TABLE IF NOT EXISTS board_shared_tasks (
    board_id   UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    task_id    UUID NOT NULL REFERENCES tasks(id)  ON DELETE CASCADE,
    shared_by  UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    shared_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (board_id, task_id)
);
CREATE INDEX IF NOT EXISTS idx_board_shared_tasks_task   ON board_shared_tasks(task_id);
CREATE INDEX IF NOT EXISTS idx_board_shared_tasks_sharer ON board_shared_tasks(shared_by);
