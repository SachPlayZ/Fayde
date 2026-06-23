-- Global full-text search: tsvector columns on tasks and comments.
-- (notes already has search_tsv from 0008.)

ALTER TABLE tasks ADD COLUMN search_tsv tsvector
  GENERATED ALWAYS AS (to_tsvector('english', title || ' ' || coalesce(description,''))) STORED;
CREATE INDEX idx_tasks_tsv ON tasks USING GIN(search_tsv);

ALTER TABLE task_comments ADD COLUMN search_tsv tsvector
  GENERATED ALWAYS AS (to_tsvector('english', coalesce(body,''))) STORED;
CREATE INDEX idx_comments_tsv ON task_comments USING GIN(search_tsv);
