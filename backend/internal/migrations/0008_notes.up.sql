-- Notes / Docs / Wiki: foldered rich docs (BlockNote JSON) with task + note links.

CREATE TABLE notes (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  parent_id  UUID REFERENCES notes(id) ON DELETE CASCADE,
  title      TEXT NOT NULL DEFAULT 'Untitled',
  body       TEXT NOT NULL DEFAULT '',  -- BlockNote document JSON
  plain      TEXT NOT NULL DEFAULT '',  -- derived plaintext for search
  is_folder  BOOLEAN NOT NULL DEFAULT false,
  icon       TEXT,
  position   DOUBLE PRECISION NOT NULL DEFAULT 0,
  archived   BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notes_user ON notes(user_id, parent_id);

ALTER TABLE notes ADD COLUMN search_tsv tsvector
  GENERATED ALWAYS AS (to_tsvector('english', title || ' ' || plain)) STORED;
CREATE INDEX idx_notes_tsv ON notes USING GIN(search_tsv);

-- doc <-> task links
CREATE TABLE note_links (
  note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  PRIMARY KEY (note_id, task_id)
);
CREATE INDEX idx_note_links_task ON note_links(task_id);

-- note <-> note backlinks resolved from [[id]] on save
CREATE TABLE note_backlinks (
  src_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  dst_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  PRIMARY KEY (src_id, dst_id)
);
CREATE INDEX idx_note_backlinks_dst ON note_backlinks(dst_id);
