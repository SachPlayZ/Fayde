CREATE TABLE leetcode_problems (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  lc_number   INT,
  slug        TEXT,
  title       TEXT        NOT NULL,
  url         TEXT,
  difficulty  TEXT        NOT NULL DEFAULT 'medium',
  topics      TEXT[]      NOT NULL DEFAULT '{}',
  notes       TEXT        NOT NULL DEFAULT '',
  solved_at   TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_leetcode_problems_user ON leetcode_problems(user_id);

CREATE TABLE leetcode_cards (
  id              UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
  problem_id      UUID             NOT NULL REFERENCES leetcode_problems(id) ON DELETE CASCADE,
  user_id         UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  stability       DOUBLE PRECISION NOT NULL DEFAULT 0,
  difficulty      DOUBLE PRECISION NOT NULL DEFAULT 0,
  elapsed_days    INT              NOT NULL DEFAULT 0,
  scheduled_days  INT              NOT NULL DEFAULT 0,
  reps            INT              NOT NULL DEFAULT 0,
  lapses          INT              NOT NULL DEFAULT 0,
  card_state      TEXT             NOT NULL DEFAULT 'new',
  due_date        TIMESTAMPTZ      NOT NULL DEFAULT now(),
  last_review     TIMESTAMPTZ,
  UNIQUE (problem_id)
);
CREATE INDEX idx_leetcode_cards_user_due ON leetcode_cards(user_id, due_date);

CREATE TABLE leetcode_review_logs (
  id              UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
  card_id         UUID             NOT NULL REFERENCES leetcode_cards(id) ON DELETE CASCADE,
  user_id         UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reviewed_at     TIMESTAMPTZ      NOT NULL DEFAULT now(),
  rating          INT              NOT NULL CHECK (rating BETWEEN 1 AND 4),
  scheduled_days  INT              NOT NULL DEFAULT 0,
  elapsed_days    INT              NOT NULL DEFAULT 0,
  stability       DOUBLE PRECISION NOT NULL DEFAULT 0,
  difficulty      DOUBLE PRECISION NOT NULL DEFAULT 0,
  card_state      TEXT             NOT NULL DEFAULT 'new'
);
CREATE INDEX idx_leetcode_review_logs_card ON leetcode_review_logs(card_id);
CREATE INDEX idx_leetcode_review_logs_user_date ON leetcode_review_logs(user_id, reviewed_at DESC);
