CREATE TABLE story (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at INTEGER,
  updated_at INTEGER,
  deleted_at INTEGER,
  title TEXT NOT NULL,
  author TEXT NOT NULL,
  description TEXT NOT NULL,
  music_style TEXT NOT NULL,
  status INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_story_deleted_at ON story (deleted_at);

CREATE TABLE chapter (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at INTEGER,
  updated_at INTEGER,
  deleted_at INTEGER,
  story_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  image_prompt TEXT NOT NULL,
  image_path TEXT NOT NULL,
  voice_path TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_chapter_deleted_at ON chapter (deleted_at);
CREATE INDEX IF NOT EXISTS idx_chapter_story_id ON chapter (story_id);