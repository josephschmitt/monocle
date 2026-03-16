// Configuration schema

export interface MonocleConfig {
  /** Review mode: 'review' for standard review, 'gate' for CI gating */
  mode: 'review' | 'gate';

  /** Glob patterns for files that must pass review before merge (gate mode) */
  gate_patterns: string[];

  /** Glob patterns for files to ignore during review */
  ignore_patterns: string[];

  /** Template for Unix socket path. Supports {{sessionId}} placeholder. */
  socket_path: string;
}
