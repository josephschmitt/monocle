// Git CLI wrapper (diff, status, file content)

import { execFile } from 'node:child_process';
import * as fs from 'node:fs';
import * as path from 'node:path';
import type { ChangeStatus } from '../types/review.js';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface FileStatusEntry {
  path: string;
  status: ChangeStatus;
  /** Original path for renames (the "from" side). */
  origPath?: string;
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

/**
 * Runs a git command via execFile and returns trimmed stdout.
 * Rejects with a descriptive error on non-zero exit.
 */
function exec(args: string[], cwd: string): Promise<string> {
  return new Promise((resolve, reject) => {
    execFile('git', args, { cwd, maxBuffer: 10 * 1024 * 1024 }, (error, stdout, stderr) => {
      if (error) {
        const msg = stderr?.trim() || error.message;
        reject(new Error(`git ${args[0]}: ${msg}`));
        return;
      }
      resolve(stdout.trimEnd());
    });
  });
}

/**
 * Maps the two-character porcelain status codes to our ChangeStatus union.
 */
function parseStatusCode(xy: string): ChangeStatus {
  const index = xy[0]!;
  const worktree = xy[1]!;

  if (index === 'R' || worktree === 'R') return 'renamed';
  if (index === 'A' || worktree === 'A') return 'added';
  if (index === 'D' || worktree === 'D') return 'deleted';
  // M, T, U, ' M', etc. all collapse to modified
  return 'modified';
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Returns the repository root for the given working directory.
 * Defaults to `process.cwd()` when no `cwd` is supplied.
 */
export async function getRepoRoot(cwd?: string): Promise<string> {
  return exec(['rev-parse', '--show-toplevel'], cwd ?? process.cwd());
}

/**
 * Returns the full SHA of the current HEAD commit.
 */
export async function getCurrentRef(repoRoot: string): Promise<string> {
  return exec(['rev-parse', 'HEAD'], repoRoot);
}

/**
 * Parses `git status --porcelain` into structured entries.
 */
export async function getStatus(repoRoot: string): Promise<FileStatusEntry[]> {
  const raw = await exec(['status', '--porcelain', '-z'], repoRoot);
  if (raw === '') return [];

  const entries: FileStatusEntry[] = [];
  // -z uses NUL as separator; split produces trailing empty element
  const parts = raw.split('\0');
  let i = 0;

  while (i < parts.length) {
    const part = parts[i]!;
    if (part === '') {
      i++;
      continue;
    }

    const xy = part.slice(0, 2);
    const filePath = part.slice(3);
    const status = parseStatusCode(xy);

    if (status === 'renamed') {
      // Renames have a second NUL-separated field: the original path
      i++;
      const origPath = parts[i];
      entries.push({ path: filePath, status, origPath });
    } else {
      entries.push({ path: filePath, status });
    }

    i++;
  }

  return entries;
}

/**
 * Returns a raw unified diff.
 *
 * - Compares `baseRef` against the working tree.
 * - Optionally scoped to a single `filePath`.
 */
export async function getDiff(
  repoRoot: string,
  baseRef: string,
  filePath?: string,
): Promise<string> {
  const args = ['diff', baseRef];
  if (filePath) {
    args.push('--', filePath);
  }
  return exec(args, repoRoot);
}

/**
 * Returns file content.
 *
 * - Without `ref`: reads from the working tree via the filesystem.
 * - With `ref`: uses `git show <ref>:<path>`.
 */
export async function getFileContent(
  repoRoot: string,
  filePath: string,
  ref?: string,
): Promise<string> {
  if (ref) {
    return exec(['show', `${ref}:${filePath}`], repoRoot);
  }
  const absPath = path.resolve(repoRoot, filePath);
  return fs.promises.readFile(absPath, 'utf-8');
}

/**
 * Extracts a 1-based inclusive line range from a working-tree file.
 */
export async function getFileLines(
  repoRoot: string,
  filePath: string,
  startLine: number,
  endLine: number,
): Promise<string> {
  const content = await fs.promises.readFile(path.resolve(repoRoot, filePath), 'utf-8');
  const lines = content.split('\n');
  // startLine and endLine are 1-based, inclusive
  return lines.slice(startLine - 1, endLine).join('\n');
}
