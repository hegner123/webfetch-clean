-- expires stores Unix epoch seconds (integer) for portable time comparison.
-- All time comparisons use unixepoch('now') to avoid ISO 8601 format mismatches
-- between Go's time.Time serialization and SQLite's datetime() function.
CREATE TABLE IF NOT EXISTS file_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    file TEXT NOT NULL,
    expires INTEGER NOT NULL,
    consumed BOOLEAN NOT NULL DEFAULT FALSE
);
