import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { execFileSync } from 'node:child_process';
import * as fs from 'node:fs';
import * as os from 'node:os';
import * as path from 'node:path';
import {
  getRepoRoot,
  getCurrentRef,
  getStatus,
  getDiff,
  getFileContent,
  getFileLines,
} from '../../src/core/git.js';

// ---------------------------------------------------------------------------
// Helpers — create a throw-away git repo for each test
// ---------------------------------------------------------------------------

let repoDir: string;

function git(...args: string[]): string {
  return execFileSync('git', args, { cwd: repoDir, encoding: 'utf-8' }).trimEnd();
}

function writeFile(name: string, content: string): void {
  const abs = path.join(repoDir, name);
  fs.mkdirSync(path.dirname(abs), { recursive: true });
  fs.writeFileSync(abs, content, 'utf-8');
}

beforeEach(() => {
  repoDir = fs.mkdtempSync(path.join(os.tmpdir(), 'monocle-git-test-'));
  git('init');
  git('config', 'user.email', 'test@test.com');
  git('config', 'user.name', 'Test');
  // Initial commit so HEAD exists
  writeFile('README.md', '# Test Repo\n');
  git('add', '.');
  git('commit', '-m', 'initial commit');
});

afterEach(() => {
  fs.rmSync(repoDir, { recursive: true, force: true });
});

// ---------------------------------------------------------------------------
// getRepoRoot
// ---------------------------------------------------------------------------

describe('getRepoRoot', () => {
  it('returns the repo root from the repo directory', async () => {
    const root = await getRepoRoot(repoDir);
    // Resolve both to handle macOS /private/var vs /var symlinks
    expect(fs.realpathSync(root)).toBe(fs.realpathSync(repoDir));
  });

  it('returns the repo root from a subdirectory', async () => {
    const sub = path.join(repoDir, 'a', 'b');
    fs.mkdirSync(sub, { recursive: true });
    const root = await getRepoRoot(sub);
    expect(fs.realpathSync(root)).toBe(fs.realpathSync(repoDir));
  });

  it('rejects for a non-repo directory', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'no-git-'));
    try {
      await expect(getRepoRoot(tmp)).rejects.toThrow();
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });
});

// ---------------------------------------------------------------------------
// getCurrentRef
// ---------------------------------------------------------------------------

describe('getCurrentRef', () => {
  it('returns a 40-char hex SHA', async () => {
    const sha = await getCurrentRef(repoDir);
    expect(sha).toMatch(/^[0-9a-f]{40}$/);
  });

  it('changes after a new commit', async () => {
    const sha1 = await getCurrentRef(repoDir);
    writeFile('new.txt', 'hello');
    git('add', 'new.txt');
    git('commit', '-m', 'second');
    const sha2 = await getCurrentRef(repoDir);
    expect(sha1).not.toBe(sha2);
  });
});

// ---------------------------------------------------------------------------
// getStatus
// ---------------------------------------------------------------------------

describe('getStatus', () => {
  it('returns empty array for clean worktree', async () => {
    const status = await getStatus(repoDir);
    expect(status).toEqual([]);
  });

  it('detects added (untracked then staged) files', async () => {
    writeFile('new.txt', 'content');
    git('add', 'new.txt');
    const status = await getStatus(repoDir);
    expect(status).toContainEqual({ path: 'new.txt', status: 'added' });
  });

  it('detects modified files', async () => {
    writeFile('README.md', '# Updated\n');
    git('add', 'README.md');
    const status = await getStatus(repoDir);
    expect(status).toContainEqual({ path: 'README.md', status: 'modified' });
  });

  it('detects deleted files', async () => {
    git('rm', 'README.md');
    const status = await getStatus(repoDir);
    expect(status).toContainEqual({ path: 'README.md', status: 'deleted' });
  });

  it('detects renamed files', async () => {
    git('mv', 'README.md', 'DOCS.md');
    const status = await getStatus(repoDir);
    const renamed = status.find((e) => e.status === 'renamed');
    expect(renamed).toBeDefined();
    expect(renamed!.path).toBe('DOCS.md');
    expect(renamed!.origPath).toBe('README.md');
  });
});

// ---------------------------------------------------------------------------
// getDiff
// ---------------------------------------------------------------------------

describe('getDiff', () => {
  it('returns empty string when nothing changed since base', async () => {
    const sha = await getCurrentRef(repoDir);
    const diff = await getDiff(repoDir, sha);
    expect(diff).toBe('');
  });

  it('returns unified diff for modified file', async () => {
    const sha = await getCurrentRef(repoDir);
    writeFile('README.md', '# Changed\n');
    const diff = await getDiff(repoDir, sha);
    expect(diff).toContain('diff --git');
    expect(diff).toContain('-# Test Repo');
    expect(diff).toContain('+# Changed');
  });

  it('scopes diff to a specific file', async () => {
    const sha = await getCurrentRef(repoDir);
    writeFile('README.md', '# Changed\n');
    writeFile('other.txt', 'new file');
    git('add', 'other.txt');
    const diff = await getDiff(repoDir, sha, 'README.md');
    expect(diff).toContain('README.md');
    expect(diff).not.toContain('other.txt');
  });
});

// ---------------------------------------------------------------------------
// getFileContent
// ---------------------------------------------------------------------------

describe('getFileContent', () => {
  it('reads working tree content without ref', async () => {
    const content = await getFileContent(repoDir, 'README.md');
    expect(content).toBe('# Test Repo\n');
  });

  it('reads content at a specific ref', async () => {
    const sha1 = await getCurrentRef(repoDir);
    writeFile('README.md', '# V2\n');
    git('add', 'README.md');
    git('commit', '-m', 'v2');

    // Current working tree is V2, but old ref should give original
    const oldContent = await getFileContent(repoDir, 'README.md', sha1);
    expect(oldContent).toBe('# Test Repo');

    const newContent = await getFileContent(repoDir, 'README.md');
    expect(newContent).toBe('# V2\n');
  });

  it('rejects for non-existent file', async () => {
    await expect(getFileContent(repoDir, 'nope.txt')).rejects.toThrow();
  });

  it('rejects for non-existent ref', async () => {
    await expect(getFileContent(repoDir, 'README.md', 'deadbeef')).rejects.toThrow();
  });
});

// ---------------------------------------------------------------------------
// getFileLines
// ---------------------------------------------------------------------------

describe('getFileLines', () => {
  it('extracts a range of lines (1-based, inclusive)', async () => {
    writeFile('multi.txt', 'line1\nline2\nline3\nline4\nline5\n');
    const lines = await getFileLines(repoDir, 'multi.txt', 2, 4);
    expect(lines).toBe('line2\nline3\nline4');
  });

  it('handles single-line extraction', async () => {
    writeFile('multi.txt', 'a\nb\nc\n');
    const lines = await getFileLines(repoDir, 'multi.txt', 2, 2);
    expect(lines).toBe('b');
  });

  it('handles range exceeding file length gracefully', async () => {
    writeFile('short.txt', 'one\ntwo\n');
    const lines = await getFileLines(repoDir, 'short.txt', 1, 100);
    expect(lines).toContain('one');
    expect(lines).toContain('two');
  });
});
