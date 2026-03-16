// Typed query functions for Monocle database operations

import type Database from 'better-sqlite3';
import type {
  ReviewSession,
  ReviewState,
  ReviewMode,
  ChangeStatus,
  ContentType,
  CommentType,
  CommentTargetType,
  SubmissionAction,
} from '../types/review.js';

// ---------------------------------------------------------------------------
// Row types (DB representation — flat, with JSON-encoded arrays)
// ---------------------------------------------------------------------------

interface SessionRow {
  id: string;
  mode: string;
  agent: string;
  repo_root: string;
  base_ref: string;
  state: string;
  gate_patterns: string | null;
  ignore_patterns: string | null;
  created_at: string;
  updated_at: string;
}

interface ChangedFileRow {
  id: string;
  session_id: string;
  path: string;
  status: string;
  reviewed: number;
  diff_data: string | null;
  updated_at: string;
}

interface ContentItemRow {
  id: string;
  session_id: string;
  title: string;
  content: string;
  content_type: string;
  reviewed: number;
  created_at: string;
}

interface CommentRow {
  id: string;
  session_id: string;
  target_type: string;
  target_ref: string;
  line_start: number;
  line_end: number;
  type: string;
  body: string;
  code_snippet: string;
  resolved: number;
  created_at: string;
}

interface SubmissionRow {
  id: string;
  session_id: string;
  action: string;
  formatted_review: string | null;
  comment_count: number | null;
  submitted_at: string;
}

// ---------------------------------------------------------------------------
// Input types for create/upsert operations
// ---------------------------------------------------------------------------

export interface CreateSessionInput {
  id: string;
  mode: ReviewMode;
  agent: string;
  repo_root: string;
  base_ref: string;
  state: ReviewState;
  gate_patterns?: string[];
  ignore_patterns?: string[];
  created_at: string;
  updated_at: string;
}

export interface UpsertChangedFileInput {
  id: string;
  session_id: string;
  path: string;
  status: ChangeStatus;
  reviewed?: boolean;
  diff_data?: string | null;
  updated_at: string;
}

export interface UpsertContentItemInput {
  id: string;
  session_id: string;
  title: string;
  content: string;
  content_type: ContentType;
  reviewed?: boolean;
  created_at: string;
}

export interface AddCommentInput {
  id: string;
  session_id: string;
  target_type: CommentTargetType;
  target_ref: string;
  line_start: number;
  line_end: number;
  type: CommentType;
  body: string;
  code_snippet: string;
  resolved?: boolean;
  created_at: string;
}

export interface UpdateCommentInput {
  body?: string;
  resolved?: boolean;
  type?: CommentType;
}

export interface AddSubmissionInput {
  id: string;
  session_id: string;
  action: SubmissionAction;
  formatted_review?: string | null;
  comment_count?: number | null;
  submitted_at: string;
}

// ---------------------------------------------------------------------------
// Output types (typed versions of rows)
// ---------------------------------------------------------------------------

export interface SessionRecord {
  id: string;
  mode: ReviewMode;
  agent: string;
  repo_root: string;
  base_ref: string;
  state: ReviewState;
  gate_patterns: string[];
  ignore_patterns: string[];
  created_at: string;
  updated_at: string;
}

export interface ChangedFileRecord {
  id: string;
  session_id: string;
  path: string;
  status: ChangeStatus;
  reviewed: boolean;
  diff_data: string | null;
  updated_at: string;
}

export interface ContentItemRecord {
  id: string;
  session_id: string;
  title: string;
  content: string;
  content_type: ContentType;
  reviewed: boolean;
  created_at: string;
}

export interface CommentRecord {
  id: string;
  session_id: string;
  target_type: CommentTargetType;
  target_ref: string;
  line_start: number;
  line_end: number;
  type: CommentType;
  body: string;
  code_snippet: string;
  resolved: boolean;
  created_at: string;
}

export interface SubmissionRecord {
  id: string;
  session_id: string;
  action: SubmissionAction;
  formatted_review: string | null;
  comment_count: number | null;
  submitted_at: string;
}

// ---------------------------------------------------------------------------
// Row → Record mappers
// ---------------------------------------------------------------------------

function toSessionRecord(row: SessionRow): SessionRecord {
  return {
    id: row.id,
    mode: row.mode as ReviewMode,
    agent: row.agent,
    repo_root: row.repo_root,
    base_ref: row.base_ref,
    state: row.state as ReviewState,
    gate_patterns: row.gate_patterns ? JSON.parse(row.gate_patterns) as string[] : [],
    ignore_patterns: row.ignore_patterns ? JSON.parse(row.ignore_patterns) as string[] : [],
    created_at: row.created_at,
    updated_at: row.updated_at,
  };
}

function toChangedFileRecord(row: ChangedFileRow): ChangedFileRecord {
  return {
    id: row.id,
    session_id: row.session_id,
    path: row.path,
    status: row.status as ChangeStatus,
    reviewed: row.reviewed === 1,
    diff_data: row.diff_data,
    updated_at: row.updated_at,
  };
}

function toContentItemRecord(row: ContentItemRow): ContentItemRecord {
  return {
    id: row.id,
    session_id: row.session_id,
    title: row.title,
    content: row.content,
    content_type: row.content_type as ContentType,
    reviewed: row.reviewed === 1,
    created_at: row.created_at,
  };
}

function toCommentRecord(row: CommentRow): CommentRecord {
  return {
    id: row.id,
    session_id: row.session_id,
    target_type: row.target_type as CommentTargetType,
    target_ref: row.target_ref,
    line_start: row.line_start,
    line_end: row.line_end,
    type: row.type as CommentType,
    body: row.body,
    code_snippet: row.code_snippet,
    resolved: row.resolved === 1,
    created_at: row.created_at,
  };
}

function toSubmissionRecord(row: SubmissionRow): SubmissionRecord {
  return {
    id: row.id,
    session_id: row.session_id,
    action: row.action as SubmissionAction,
    formatted_review: row.formatted_review,
    comment_count: row.comment_count,
    submitted_at: row.submitted_at,
  };
}

// ---------------------------------------------------------------------------
// Session queries
// ---------------------------------------------------------------------------

export function createSession(db: Database.Database, session: CreateSessionInput): SessionRecord {
  const stmt = db.prepare(`
    INSERT INTO sessions (id, mode, agent, repo_root, base_ref, state, gate_patterns, ignore_patterns, created_at, updated_at)
    VALUES (@id, @mode, @agent, @repo_root, @base_ref, @state, @gate_patterns, @ignore_patterns, @created_at, @updated_at)
  `);

  stmt.run({
    id: session.id,
    mode: session.mode,
    agent: session.agent,
    repo_root: session.repo_root,
    base_ref: session.base_ref,
    state: session.state,
    gate_patterns: session.gate_patterns ? JSON.stringify(session.gate_patterns) : null,
    ignore_patterns: session.ignore_patterns ? JSON.stringify(session.ignore_patterns) : null,
    created_at: session.created_at,
    updated_at: session.updated_at,
  });

  return getSession(db, session.id)!;
}

export function getSession(db: Database.Database, id: string): SessionRecord | undefined {
  const row = db.prepare('SELECT * FROM sessions WHERE id = ?').get(id) as SessionRow | undefined;
  return row ? toSessionRecord(row) : undefined;
}

export function updateSessionState(db: Database.Database, id: string, state: ReviewState): void {
  const now = new Date().toISOString();
  db.prepare('UPDATE sessions SET state = ?, updated_at = ? WHERE id = ?').run(state, now, id);
}

export function listSessions(db: Database.Database, repoRoot?: string): SessionRecord[] {
  let rows: SessionRow[];
  if (repoRoot) {
    rows = db.prepare('SELECT * FROM sessions WHERE repo_root = ? ORDER BY created_at DESC').all(repoRoot) as SessionRow[];
  } else {
    rows = db.prepare('SELECT * FROM sessions ORDER BY created_at DESC').all() as SessionRow[];
  }
  return rows.map(toSessionRecord);
}

export function getLatestSession(db: Database.Database, repoRoot?: string): SessionRecord | undefined {
  let row: SessionRow | undefined;
  if (repoRoot) {
    row = db.prepare('SELECT * FROM sessions WHERE repo_root = ? ORDER BY created_at DESC LIMIT 1').get(repoRoot) as SessionRow | undefined;
  } else {
    row = db.prepare('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 1').get() as SessionRow | undefined;
  }
  return row ? toSessionRecord(row) : undefined;
}

// ---------------------------------------------------------------------------
// Changed file queries
// ---------------------------------------------------------------------------

export function upsertChangedFile(db: Database.Database, sessionId: string, file: Omit<UpsertChangedFileInput, 'session_id'>): ChangedFileRecord {
  const stmt = db.prepare(`
    INSERT INTO changed_files (id, session_id, path, status, reviewed, diff_data, updated_at)
    VALUES (@id, @session_id, @path, @status, @reviewed, @diff_data, @updated_at)
    ON CONFLICT(id) DO UPDATE SET
      path = excluded.path,
      status = excluded.status,
      reviewed = excluded.reviewed,
      diff_data = excluded.diff_data,
      updated_at = excluded.updated_at
  `);

  stmt.run({
    id: file.id,
    session_id: sessionId,
    path: file.path,
    status: file.status,
    reviewed: file.reviewed ? 1 : 0,
    diff_data: file.diff_data ?? null,
    updated_at: file.updated_at,
  });

  const row = db.prepare('SELECT * FROM changed_files WHERE id = ?').get(file.id) as ChangedFileRow;
  return toChangedFileRecord(row);
}

export function getChangedFiles(db: Database.Database, sessionId: string): ChangedFileRecord[] {
  const rows = db.prepare('SELECT * FROM changed_files WHERE session_id = ? ORDER BY path').all(sessionId) as ChangedFileRow[];
  return rows.map(toChangedFileRecord);
}

// ---------------------------------------------------------------------------
// Content item queries
// ---------------------------------------------------------------------------

export function upsertContentItem(db: Database.Database, sessionId: string, item: Omit<UpsertContentItemInput, 'session_id'>): ContentItemRecord {
  const stmt = db.prepare(`
    INSERT INTO content_items (id, session_id, title, content, content_type, reviewed, created_at)
    VALUES (@id, @session_id, @title, @content, @content_type, @reviewed, @created_at)
    ON CONFLICT(id) DO UPDATE SET
      title = excluded.title,
      content = excluded.content,
      content_type = excluded.content_type,
      reviewed = excluded.reviewed,
      created_at = excluded.created_at
  `);

  stmt.run({
    id: item.id,
    session_id: sessionId,
    title: item.title,
    content: item.content,
    content_type: item.content_type,
    reviewed: item.reviewed ? 1 : 0,
    created_at: item.created_at,
  });

  const row = db.prepare('SELECT * FROM content_items WHERE id = ?').get(item.id) as ContentItemRow;
  return toContentItemRecord(row);
}

export function getContentItems(db: Database.Database, sessionId: string): ContentItemRecord[] {
  const rows = db.prepare('SELECT * FROM content_items WHERE session_id = ? ORDER BY created_at').all(sessionId) as ContentItemRow[];
  return rows.map(toContentItemRecord);
}

// ---------------------------------------------------------------------------
// Comment queries
// ---------------------------------------------------------------------------

export function addComment(db: Database.Database, comment: AddCommentInput): CommentRecord {
  const stmt = db.prepare(`
    INSERT INTO comments (id, session_id, target_type, target_ref, line_start, line_end, type, body, code_snippet, resolved, created_at)
    VALUES (@id, @session_id, @target_type, @target_ref, @line_start, @line_end, @type, @body, @code_snippet, @resolved, @created_at)
  `);

  stmt.run({
    id: comment.id,
    session_id: comment.session_id,
    target_type: comment.target_type,
    target_ref: comment.target_ref,
    line_start: comment.line_start,
    line_end: comment.line_end,
    type: comment.type,
    body: comment.body,
    code_snippet: comment.code_snippet,
    resolved: comment.resolved ? 1 : 0,
    created_at: comment.created_at,
  });

  const row = db.prepare('SELECT * FROM comments WHERE id = ?').get(comment.id) as CommentRow;
  return toCommentRecord(row);
}

export function getComments(db: Database.Database, sessionId: string): CommentRecord[] {
  const rows = db.prepare('SELECT * FROM comments WHERE session_id = ? ORDER BY created_at').all(sessionId) as CommentRow[];
  return rows.map(toCommentRecord);
}

export function getCommentsByTarget(
  db: Database.Database,
  sessionId: string,
  targetType: CommentTargetType,
  targetRef: string,
): CommentRecord[] {
  const rows = db.prepare(
    'SELECT * FROM comments WHERE session_id = ? AND target_type = ? AND target_ref = ? ORDER BY created_at',
  ).all(sessionId, targetType, targetRef) as CommentRow[];
  return rows.map(toCommentRecord);
}

export function updateComment(db: Database.Database, id: string, updates: UpdateCommentInput): CommentRecord | undefined {
  const setClauses: string[] = [];
  const params: Record<string, unknown> = { id };

  if (updates.body !== undefined) {
    setClauses.push('body = @body');
    params['body'] = updates.body;
  }
  if (updates.resolved !== undefined) {
    setClauses.push('resolved = @resolved');
    params['resolved'] = updates.resolved ? 1 : 0;
  }
  if (updates.type !== undefined) {
    setClauses.push('type = @type');
    params['type'] = updates.type;
  }

  if (setClauses.length === 0) {
    const row = db.prepare('SELECT * FROM comments WHERE id = ?').get(id) as CommentRow | undefined;
    return row ? toCommentRecord(row) : undefined;
  }

  db.prepare(`UPDATE comments SET ${setClauses.join(', ')} WHERE id = @id`).run(params);

  const row = db.prepare('SELECT * FROM comments WHERE id = ?').get(id) as CommentRow | undefined;
  return row ? toCommentRecord(row) : undefined;
}

export function deleteComment(db: Database.Database, id: string): boolean {
  const result = db.prepare('DELETE FROM comments WHERE id = ?').run(id);
  return result.changes > 0;
}

// ---------------------------------------------------------------------------
// Submission queries
// ---------------------------------------------------------------------------

export function addSubmission(db: Database.Database, submission: AddSubmissionInput): SubmissionRecord {
  const stmt = db.prepare(`
    INSERT INTO review_submissions (id, session_id, action, formatted_review, comment_count, submitted_at)
    VALUES (@id, @session_id, @action, @formatted_review, @comment_count, @submitted_at)
  `);

  stmt.run({
    id: submission.id,
    session_id: submission.session_id,
    action: submission.action,
    formatted_review: submission.formatted_review ?? null,
    comment_count: submission.comment_count ?? null,
    submitted_at: submission.submitted_at,
  });

  const row = db.prepare('SELECT * FROM review_submissions WHERE id = ?').get(submission.id) as SubmissionRow;
  return toSubmissionRecord(row);
}

export function getSubmissions(db: Database.Database, sessionId: string): SubmissionRecord[] {
  const rows = db.prepare('SELECT * FROM review_submissions WHERE session_id = ? ORDER BY submitted_at').all(sessionId) as SubmissionRow[];
  return rows.map(toSubmissionRecord);
}
