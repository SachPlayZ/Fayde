-- Delete unused kanban "projects" feature, repurpose the showcase entries
-- table as the new "projects" (public portfolio) feature, and drop the
-- dedicated showcase bearer-token type in favor of existing api_tokens.

ALTER TABLE tasks DROP COLUMN IF EXISTS project_id;
DROP TABLE IF EXISTS projects;

ALTER TABLE showcase_entries RENAME TO projects;
ALTER INDEX idx_showcase_entries_user RENAME TO idx_projects_user;

DROP TABLE IF EXISTS showcase_tokens;

ALTER TABLE tasks ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
