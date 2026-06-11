DROP TABLE IF EXISTS task_attachments;
DROP INDEX IF EXISTS idx_attachments_task;

DROP TABLE IF EXISTS activity_logs;
DROP INDEX IF EXISTS idx_activity_logs_task;

ALTER TABLE users DROP COLUMN IF EXISTS role;
