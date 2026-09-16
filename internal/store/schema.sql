CREATE TABLE IF NOT EXISTS activities (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
    id               INTEGER PRIMARY KEY,
    activity_id      INTEGER NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    name             TEXT NOT NULL,
    expected_minutes INTEGER NOT NULL DEFAULT 0,
    created_at       TEXT NOT NULL,
    UNIQUE (activity_id, name)
);

CREATE TABLE IF NOT EXISTS sessions (
    id               INTEGER PRIMARY KEY,
    project_id       INTEGER NOT NULL REFERENCES projects(id),
    note             TEXT NOT NULL DEFAULT '',
    expected_minutes INTEGER NOT NULL DEFAULT 0,
    start_at         TEXT NOT NULL,
    end_at           TEXT,
    CHECK (end_at IS NULL OR end_at >= start_at)
);

CREATE INDEX IF NOT EXISTS idx_sessions_start ON sessions(start_at);
CREATE INDEX IF NOT EXISTS idx_sessions_end   ON sessions(end_at);
