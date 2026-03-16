import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import Database from 'better-sqlite3';
import { initDatabase } from '../../src/db/schema.js';
import {
  createSession,
  getSession,
  updateSessionState,
  listSessions,
  getLatestSession,
  upsertChangedFile,
  getChangedFiles,
  upsertContentItem,
  getContentItems,
  addComment,
  getComments,
  getCommentsByTarget,
  updateComment,
  deleteComment,
  addSubmission,
  getSubmissions,
} from '../../src/db/queries.js';
import type Database_Type from 'better-sqlite3';

let db: Database_Type.Database;

beforeEach(() => {
  // Use in-memory database for tests
  db = initDatabase(':memory:');
});

afterEach(() => {
  db.close();
});

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

describe('sessions', () => {
  const baseSession = {
    id: 'sess-1',
    mode: 'review' as const,
    agent: 'claude',
    repo_root: '/home/user/project',
    base_ref: 'HEAD',
    state: 'idle' as const,
    gate_patterns: ['src/**'],
    ignore_patterns: ['node_modules/**'],
    created_at: '2026-01-01T00:00:00.000Z',
    updated_at: '2026-01-01T00:00:00.000Z',
  };

  it('creates and retrieves a session', () => {
    const created = createSession(db, baseSession);

    expect(created.id).toBe('sess-1');
    expect(created.mode).toBe('review');
    expect(created.agent).toBe('claude');
    expect(created.state).toBe('idle');
    expect(created.gate_patterns).toEqual(['src/**']);
    expect(created.ignore_patterns).toEqual(['node_modules/**']);

    const fetched = getSession(db, 'sess-1');
    expect(fetched).toEqual(created);
  });

  it('returns undefined for non-existent session', () => {
    expect(getSession(db, 'non-existent')).toBeUndefined();
  });

  it('updates session state', () => {
    createSession(db, baseSession);
    updateSessionState(db, 'sess-1', 'reviewing');

    const fetched = getSession(db, 'sess-1');
    expect(fetched!.state).toBe('reviewing');
    // updated_at should have changed
    expect(fetched!.updated_at).not.toBe(baseSession.updated_at);
  });

  it('lists all sessions ordered by created_at desc', () => {
    createSession(db, { ...baseSession, id: 'sess-1', created_at: '2026-01-01T00:00:00.000Z' });
    createSession(db, { ...baseSession, id: 'sess-2', created_at: '2026-01-02T00:00:00.000Z' });
    createSession(db, { ...baseSession, id: 'sess-3', created_at: '2026-01-03T00:00:00.000Z' });

    const all = listSessions(db);
    expect(all).toHaveLength(3);
    expect(all[0]!.id).toBe('sess-3');
    expect(all[2]!.id).toBe('sess-1');
  });

  it('lists sessions filtered by repo root', () => {
    createSession(db, { ...baseSession, id: 'sess-1', repo_root: '/project-a' });
    createSession(db, { ...baseSession, id: 'sess-2', repo_root: '/project-b' });
    createSession(db, { ...baseSession, id: 'sess-3', repo_root: '/project-a' });

    const filtered = listSessions(db, '/project-a');
    expect(filtered).toHaveLength(2);
    expect(filtered.every(s => s.repo_root === '/project-a')).toBe(true);
  });

  it('gets the latest session', () => {
    createSession(db, { ...baseSession, id: 'sess-1', created_at: '2026-01-01T00:00:00.000Z' });
    createSession(db, { ...baseSession, id: 'sess-2', created_at: '2026-01-03T00:00:00.000Z' });
    createSession(db, { ...baseSession, id: 'sess-3', created_at: '2026-01-02T00:00:00.000Z' });

    const latest = getLatestSession(db);
    expect(latest!.id).toBe('sess-2');
  });

  it('gets the latest session filtered by repo root', () => {
    createSession(db, { ...baseSession, id: 'sess-1', repo_root: '/project-a', created_at: '2026-01-03T00:00:00.000Z' });
    createSession(db, { ...baseSession, id: 'sess-2', repo_root: '/project-b', created_at: '2026-01-04T00:00:00.000Z' });

    const latest = getLatestSession(db, '/project-a');
    expect(latest!.id).toBe('sess-1');
  });

  it('returns undefined when no sessions exist', () => {
    expect(getLatestSession(db)).toBeUndefined();
  });

  it('handles null gate_patterns and ignore_patterns', () => {
    const session = { ...baseSession, id: 'sess-null', gate_patterns: undefined, ignore_patterns: undefined };
    const created = createSession(db, session);
    expect(created.gate_patterns).toEqual([]);
    expect(created.ignore_patterns).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// Changed files
// ---------------------------------------------------------------------------

describe('changed files', () => {
  beforeEach(() => {
    createSession(db, {
      id: 'sess-1',
      mode: 'review',
      agent: 'claude',
      repo_root: '/project',
      base_ref: 'HEAD',
      state: 'reviewing',
      created_at: '2026-01-01T00:00:00.000Z',
      updated_at: '2026-01-01T00:00:00.000Z',
    });
  });

  it('upserts and retrieves changed files', () => {
    const file = upsertChangedFile(db, 'sess-1', {
      id: 'file-1',
      path: 'src/index.ts',
      status: 'modified',
      reviewed: false,
      diff_data: '{"hunks":[]}',
      updated_at: '2026-01-01T00:00:00.000Z',
    });

    expect(file.id).toBe('file-1');
    expect(file.path).toBe('src/index.ts');
    expect(file.status).toBe('modified');
    expect(file.reviewed).toBe(false);

    const files = getChangedFiles(db, 'sess-1');
    expect(files).toHaveLength(1);
    expect(files[0]).toEqual(file);
  });

  it('upsert updates existing file', () => {
    upsertChangedFile(db, 'sess-1', {
      id: 'file-1',
      path: 'src/index.ts',
      status: 'modified',
      updated_at: '2026-01-01T00:00:00.000Z',
    });

    upsertChangedFile(db, 'sess-1', {
      id: 'file-1',
      path: 'src/index.ts',
      status: 'deleted',
      reviewed: true,
      updated_at: '2026-01-02T00:00:00.000Z',
    });

    const files = getChangedFiles(db, 'sess-1');
    expect(files).toHaveLength(1);
    expect(files[0]!.status).toBe('deleted');
    expect(files[0]!.reviewed).toBe(true);
  });

  it('returns files ordered by path', () => {
    upsertChangedFile(db, 'sess-1', { id: 'f-2', path: 'src/b.ts', status: 'added', updated_at: '2026-01-01T00:00:00.000Z' });
    upsertChangedFile(db, 'sess-1', { id: 'f-1', path: 'src/a.ts', status: 'added', updated_at: '2026-01-01T00:00:00.000Z' });

    const files = getChangedFiles(db, 'sess-1');
    expect(files[0]!.path).toBe('src/a.ts');
    expect(files[1]!.path).toBe('src/b.ts');
  });
});

// ---------------------------------------------------------------------------
// Content items
// ---------------------------------------------------------------------------

describe('content items', () => {
  beforeEach(() => {
    createSession(db, {
      id: 'sess-1',
      mode: 'review',
      agent: 'claude',
      repo_root: '/project',
      base_ref: 'HEAD',
      state: 'reviewing',
      created_at: '2026-01-01T00:00:00.000Z',
      updated_at: '2026-01-01T00:00:00.000Z',
    });
  });

  it('upserts and retrieves content items', () => {
    const item = upsertContentItem(db, 'sess-1', {
      id: 'ci-1',
      title: 'Architecture Plan',
      content: '# Plan\nDo the thing.',
      content_type: 'markdown',
      reviewed: false,
      created_at: '2026-01-01T00:00:00.000Z',
    });

    expect(item.id).toBe('ci-1');
    expect(item.title).toBe('Architecture Plan');
    expect(item.content_type).toBe('markdown');

    const items = getContentItems(db, 'sess-1');
    expect(items).toHaveLength(1);
    expect(items[0]).toEqual(item);
  });

  it('upsert by explicit ID replaces content but preserves ID', () => {
    upsertContentItem(db, 'sess-1', {
      id: 'ci-1',
      title: 'Original Title',
      content: 'Original content',
      content_type: 'markdown',
      created_at: '2026-01-01T00:00:00.000Z',
    });

    upsertContentItem(db, 'sess-1', {
      id: 'ci-1',
      title: 'Updated Title',
      content: 'Updated content',
      content_type: 'text',
      created_at: '2026-01-02T00:00:00.000Z',
    });

    const items = getContentItems(db, 'sess-1');
    expect(items).toHaveLength(1);
    expect(items[0]!.id).toBe('ci-1');
    expect(items[0]!.title).toBe('Updated Title');
    expect(items[0]!.content).toBe('Updated content');
    expect(items[0]!.content_type).toBe('text');
  });
});

// ---------------------------------------------------------------------------
// Comments
// ---------------------------------------------------------------------------

describe('comments', () => {
  beforeEach(() => {
    createSession(db, {
      id: 'sess-1',
      mode: 'review',
      agent: 'claude',
      repo_root: '/project',
      base_ref: 'HEAD',
      state: 'reviewing',
      created_at: '2026-01-01T00:00:00.000Z',
      updated_at: '2026-01-01T00:00:00.000Z',
    });
  });

  const baseComment = {
    id: 'cmt-1',
    session_id: 'sess-1',
    target_type: 'file' as const,
    target_ref: 'src/index.ts',
    line_start: 10,
    line_end: 15,
    type: 'issue' as const,
    body: 'This needs fixing',
    code_snippet: 'const x = 1;',
    resolved: false,
    created_at: '2026-01-01T00:00:00.000Z',
  };

  it('adds and retrieves comments', () => {
    const comment = addComment(db, baseComment);

    expect(comment.id).toBe('cmt-1');
    expect(comment.target_type).toBe('file');
    expect(comment.target_ref).toBe('src/index.ts');
    expect(comment.resolved).toBe(false);

    const comments = getComments(db, 'sess-1');
    expect(comments).toHaveLength(1);
    expect(comments[0]).toEqual(comment);
  });

  it('gets comments by target', () => {
    addComment(db, baseComment);
    addComment(db, { ...baseComment, id: 'cmt-2', target_type: 'content', target_ref: 'ci-1' });
    addComment(db, { ...baseComment, id: 'cmt-3', target_ref: 'src/other.ts' });

    const fileComments = getCommentsByTarget(db, 'sess-1', 'file', 'src/index.ts');
    expect(fileComments).toHaveLength(1);
    expect(fileComments[0]!.id).toBe('cmt-1');

    const contentComments = getCommentsByTarget(db, 'sess-1', 'content', 'ci-1');
    expect(contentComments).toHaveLength(1);
    expect(contentComments[0]!.id).toBe('cmt-2');
  });

  it('updates a comment', () => {
    addComment(db, baseComment);

    const updated = updateComment(db, 'cmt-1', { body: 'Updated body', resolved: true });
    expect(updated!.body).toBe('Updated body');
    expect(updated!.resolved).toBe(true);
    // Other fields unchanged
    expect(updated!.type).toBe('issue');
  });

  it('updates comment type', () => {
    addComment(db, baseComment);

    const updated = updateComment(db, 'cmt-1', { type: 'suggestion' });
    expect(updated!.type).toBe('suggestion');
  });

  it('returns undefined when updating non-existent comment', () => {
    expect(updateComment(db, 'non-existent', { body: 'test' })).toBeUndefined();
  });

  it('deletes a comment', () => {
    addComment(db, baseComment);

    expect(deleteComment(db, 'cmt-1')).toBe(true);
    expect(getComments(db, 'sess-1')).toHaveLength(0);
  });

  it('returns false when deleting non-existent comment', () => {
    expect(deleteComment(db, 'non-existent')).toBe(false);
  });

  it('handles empty update gracefully', () => {
    addComment(db, baseComment);
    const result = updateComment(db, 'cmt-1', {});
    expect(result!.body).toBe(baseComment.body);
  });
});

// ---------------------------------------------------------------------------
// Submissions
// ---------------------------------------------------------------------------

describe('submissions', () => {
  beforeEach(() => {
    createSession(db, {
      id: 'sess-1',
      mode: 'review',
      agent: 'claude',
      repo_root: '/project',
      base_ref: 'HEAD',
      state: 'submitted',
      created_at: '2026-01-01T00:00:00.000Z',
      updated_at: '2026-01-01T00:00:00.000Z',
    });
  });

  it('adds and retrieves submissions', () => {
    const submission = addSubmission(db, {
      id: 'sub-1',
      session_id: 'sess-1',
      action: 'approve',
      formatted_review: '# Review\nLGTM',
      comment_count: 3,
      submitted_at: '2026-01-01T00:00:00.000Z',
    });

    expect(submission.id).toBe('sub-1');
    expect(submission.action).toBe('approve');
    expect(submission.comment_count).toBe(3);

    const submissions = getSubmissions(db, 'sess-1');
    expect(submissions).toHaveLength(1);
    expect(submissions[0]).toEqual(submission);
  });

  it('handles null formatted_review and comment_count', () => {
    const submission = addSubmission(db, {
      id: 'sub-1',
      session_id: 'sess-1',
      action: 'comment',
      submitted_at: '2026-01-01T00:00:00.000Z',
    });

    expect(submission.formatted_review).toBeNull();
    expect(submission.comment_count).toBeNull();
  });

  it('returns submissions ordered by submitted_at', () => {
    addSubmission(db, { id: 'sub-2', session_id: 'sess-1', action: 'comment', submitted_at: '2026-01-02T00:00:00.000Z' });
    addSubmission(db, { id: 'sub-1', session_id: 'sess-1', action: 'approve', submitted_at: '2026-01-01T00:00:00.000Z' });

    const submissions = getSubmissions(db, 'sess-1');
    expect(submissions[0]!.id).toBe('sub-1');
    expect(submissions[1]!.id).toBe('sub-2');
  });
});

// ---------------------------------------------------------------------------
// Schema / initDatabase
// ---------------------------------------------------------------------------

describe('initDatabase', () => {
  it('creates all tables in a fresh database', () => {
    const freshDb = initDatabase(':memory:');
    const tables = freshDb.prepare(
      "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"
    ).all() as { name: string }[];

    const tableNames = tables.map(t => t.name);
    expect(tableNames).toContain('sessions');
    expect(tableNames).toContain('changed_files');
    expect(tableNames).toContain('content_items');
    expect(tableNames).toContain('comments');
    expect(tableNames).toContain('review_submissions');

    freshDb.close();
  });

  it('is idempotent — calling twice does not error', () => {
    const freshDb = initDatabase(':memory:');
    // Calling exec with CREATE TABLE IF NOT EXISTS again should be fine
    expect(() => initDatabase(':memory:')).not.toThrow();
    freshDb.close();
  });
});
