-- Local scores and settings for Termsnake.
CREATE TABLE IF NOT EXISTS scores (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    player_name TEXT    NOT NULL DEFAULT 'PLAYER',
    score       INTEGER NOT NULL CHECK (score >= 0),
    difficulty  TEXT    NOT NULL CHECK (difficulty IN ('easy','normal','hard')),
    snake_len   INTEGER NOT NULL,
    duration_ms INTEGER NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_scores_rank ON scores (difficulty, score DESC);

CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
