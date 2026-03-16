// SQLite schema and database initialization

import Database from 'better-sqlite3';
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as os from 'node:os';

function getDefaultDbPath(): string {
  const xdgData = process.env['XDG_DATA_HOME'] || path.join(os.homedir(), '.local', 'share');
  return path.join(xdgData, 'monocle', 'monocle.db');
}

const SCHEMA_SQL = `
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  mode TEXT NOT NULL,
  agent TEXT NOT NULL,
  repo_root TEXT NOT NULL,
  base_ref TEXT NOT NULL,
  state TEXT NOT NULL,
  gate_patterns TEXT,
  ignore_patterns TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS changed_files (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES sessions(id),
  path TEXT NOT NULL,
  status TEXT NOT NULL,
  reviewed INTEGER DEFAULT 0,
  diff_data TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS content_items (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES sessions(id),
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  content_type TEXT NOT NULL,
  reviewed INTEGER DEFAULT 0,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS comments (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES sessions(id),
  target_type TEXT NOT NULL,
  target_ref TEXT NOT NULL,
  line_start INTEGER NOT NULL,
  line_end INTEGER NOT NULL,
  type TEXT NOT NULL,
  body TEXT NOT NULL,
  code_snippet TEXT NOT NULL,
  resolved INTEGER DEFAULT 0,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS review_submissions (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES sessions(id),
  action TEXT NOT NULL,
  formatted_review TEXT,
  comment_count INTEGER,
  submitted_at TEXT NOT NULL
);
`;

/**
 * Creates or opens a SQLite database at the given path (or XDG default),
 * creating tables if they don't exist.
 */
export function initDatabase(dbPath?: string): Database.Database {
  const resolvedPath = dbPath ?? getDefaultDbPath();

  // Ensure parent directory exists
  const dir = path.dirname(resolvedPath);
  fs.mkdirSync(dir, { recursive: true });

  const db = new Database(resolvedPath);

  // Enable WAL mode for better concurrent read performance
  db.pragma('journal_mode = WAL');
  db.pragma('foreign_keys = ON');

  db.exec(SCHEMA_SQL);

  return db;
}
