-- Add migration script here

CREATE TABLE IF NOT EXISTS posts
(
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  content        TEXT    NOT NULL,
  files          TEXT,
  color          TEXT,
  shared         Boolean NOT NULL DEFAULT FALSE,
  deleted_at     BIGINT,
  created_at     BIGINT  NOT NULL,
  updated_at     BIGINT  NOT NULL,
  parent_id      INTEGER,
  children_count INTEGER NOT NULL DEFAULT 0,
  FOREIGN KEY (parent_id) REFERENCES posts (id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS tags
(
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  name       TEXT    NOT NULL UNIQUE,
  sticky     BOOLEAN NOT NULL DEFAULT FALSE,
  created_at BIGINT  NOT NULL,
  updated_at BIGINT  NOT NULL
);

CREATE TABLE IF NOT EXISTS tag_post_assoc
(
  tag_id  INTEGER NOT NULL,
  post_id INTEGER NOT NULL,
  FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE,
  FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
  UNIQUE (tag_id, post_id)
);
