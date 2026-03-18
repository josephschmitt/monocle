package db

import (
	"database/sql"
	"fmt"
)

const schemaVersion = 1

const dropSQL = `
DROP TABLE IF EXISTS review_submissions;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS content_items;
DROP TABLE IF EXISTS changed_files;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS schema_version;
`

const schemaSQL = `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	agent TEXT NOT NULL,
	agent_status TEXT NOT NULL DEFAULT 'idle',
	repo_root TEXT NOT NULL,
	base_ref TEXT NOT NULL,
	ignore_patterns TEXT NOT NULL DEFAULT '[]',
	review_round INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS changed_files (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	path TEXT NOT NULL,
	status TEXT NOT NULL,
	reviewed INTEGER NOT NULL DEFAULT 0,
	UNIQUE(session_id, path)
);

CREATE TABLE IF NOT EXISTS content_items (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	title TEXT NOT NULL,
	content TEXT NOT NULL,
	content_type TEXT NOT NULL DEFAULT 'text',
	reviewed INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS comments (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	target_type TEXT NOT NULL,
	target_ref TEXT NOT NULL,
	line_start INTEGER,
	line_end INTEGER,
	type TEXT NOT NULL,
	body TEXT NOT NULL,
	code_snippet TEXT NOT NULL DEFAULT '',
	resolved INTEGER NOT NULL DEFAULT 0,
	outdated INTEGER NOT NULL DEFAULT 0,
	review_round INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS review_submissions (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	action TEXT NOT NULL,
	formatted_review TEXT NOT NULL,
	comment_count INTEGER NOT NULL DEFAULT 0,
	review_round INTEGER NOT NULL DEFAULT 1,
	submitted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_changed_files_session ON changed_files(session_id);
CREATE INDEX IF NOT EXISTS idx_content_items_session ON content_items(session_id);
CREATE INDEX IF NOT EXISTS idx_comments_session ON comments(session_id);
CREATE INDEX IF NOT EXISTS idx_comments_target ON comments(target_type, target_ref);
CREATE INDEX IF NOT EXISTS idx_review_submissions_session ON review_submissions(session_id);
`

// Migrate checks the schema version and applies migrations as needed.
func Migrate(db *sql.DB) error {
	// Check if schema_version table exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'").Scan(&count)
	if err != nil {
		return fmt.Errorf("check schema_version: %w", err)
	}

	if count == 0 {
		// Fresh database — apply full schema
		if _, err := db.Exec(schemaSQL); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}
		if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (?)", schemaVersion); err != nil {
			return fmt.Errorf("set schema version: %w", err)
		}
		return nil
	}

	// Check current version
	var currentVersion int
	err = db.QueryRow("SELECT version FROM schema_version LIMIT 1").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	if currentVersion > schemaVersion {
		return fmt.Errorf("database schema version %d is newer than supported version %d", currentVersion, schemaVersion)
	}

	if currentVersion == schemaVersion {
		// Verify schema integrity — check that key columns exist.
		var colCount int
		err = db.QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name = 'agent_status'",
		).Scan(&colCount)
		if err != nil || colCount == 1 {
			return nil // schema looks good
		}
		// Schema is stale — fall through to recreate.
	}

	// Drop and recreate (safe during pre-release development).
	if _, err := db.Exec(dropSQL); err != nil {
		return fmt.Errorf("drop old schema: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (?)", schemaVersion); err != nil {
		return fmt.Errorf("set schema version: %w", err)
	}
	return nil
}
