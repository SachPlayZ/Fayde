DROP INDEX IF EXISTS idx_comments_tsv;
ALTER TABLE task_comments DROP COLUMN IF EXISTS search_tsv;
DROP INDEX IF EXISTS idx_tasks_tsv;
ALTER TABLE tasks DROP COLUMN IF EXISTS search_tsv;
