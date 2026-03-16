// Configuration loading and validation

import * as fs from 'node:fs';
import * as path from 'node:path';
import * as os from 'node:os';
import type { MonocleConfig } from '../types/config.js';

export const DEFAULT_CONFIG: MonocleConfig = {
  mode: 'review',
  gate_patterns: [],
  ignore_patterns: [
    'node_modules/**',
    'dist/**',
    'build/**',
    '.git/**',
    'coverage/**',
    '*.min.js',
    '*.min.css',
    'package-lock.json',
    'yarn.lock',
    'pnpm-lock.yaml',
  ],
  socket_path: path.join(os.tmpdir(), 'monocle-{{sessionId}}.sock'),
};

const VALID_KEYS = new Set<string>(['mode', 'gate_patterns', 'ignore_patterns', 'socket_path']);

function getGlobalConfigPath(): string {
  const xdgConfig = process.env['XDG_CONFIG_HOME'] || path.join(os.homedir(), '.config');
  return path.join(xdgConfig, 'monocle', 'config.json');
}

function getProjectConfigPath(projectRoot: string): string {
  return path.join(projectRoot, '.monocle', 'config.json');
}

function readJsonFile(filePath: string): Record<string, unknown> | null {
  try {
    const content = fs.readFileSync(filePath, 'utf-8');
    return JSON.parse(content) as Record<string, unknown>;
  } catch {
    return null;
  }
}

function validateConfig(raw: Record<string, unknown>, source: string): Partial<MonocleConfig> {
  const result: Partial<MonocleConfig> = {};

  for (const key of Object.keys(raw)) {
    if (!VALID_KEYS.has(key)) {
      console.warn(`monocle: unknown config key "${key}" in ${source}`);
    }
  }

  if ('mode' in raw) {
    if (raw['mode'] === 'review' || raw['mode'] === 'gate') {
      result.mode = raw['mode'];
    } else {
      console.warn(`monocle: invalid mode "${String(raw['mode'])}" in ${source}, expected "review" or "gate"`);
    }
  }

  if ('gate_patterns' in raw) {
    if (Array.isArray(raw['gate_patterns']) && raw['gate_patterns'].every((p: unknown) => typeof p === 'string')) {
      result.gate_patterns = raw['gate_patterns'] as string[];
    } else {
      console.warn(`monocle: invalid gate_patterns in ${source}, expected string array`);
    }
  }

  if ('ignore_patterns' in raw) {
    if (Array.isArray(raw['ignore_patterns']) && raw['ignore_patterns'].every((p: unknown) => typeof p === 'string')) {
      result.ignore_patterns = raw['ignore_patterns'] as string[];
    } else {
      console.warn(`monocle: invalid ignore_patterns in ${source}, expected string array`);
    }
  }

  if ('socket_path' in raw) {
    if (typeof raw['socket_path'] === 'string') {
      result.socket_path = raw['socket_path'];
    } else {
      console.warn(`monocle: invalid socket_path in ${source}, expected string`);
    }
  }

  return result;
}

/**
 * Loads and merges Monocle configuration from:
 * 1. Built-in defaults
 * 2. Global config (~/.config/monocle/config.json, XDG compliant)
 * 3. Project config (.monocle/config.json, overrides global)
 */
export function loadConfig(projectRoot?: string): MonocleConfig {
  let config: MonocleConfig = { ...DEFAULT_CONFIG, ignore_patterns: [...DEFAULT_CONFIG.ignore_patterns] };

  // Layer 2: Global config
  const globalPath = getGlobalConfigPath();
  const globalRaw = readJsonFile(globalPath);
  if (globalRaw) {
    const validated = validateConfig(globalRaw, globalPath);
    config = { ...config, ...validated };
  }

  // Layer 3: Project config
  if (projectRoot) {
    const projectPath = getProjectConfigPath(projectRoot);
    const projectRaw = readJsonFile(projectPath);
    if (projectRaw) {
      const validated = validateConfig(projectRaw, projectPath);
      config = { ...config, ...validated };
    }
  }

  return config;
}

/**
 * Resolves socket path by replacing {{sessionId}} in the template.
 */
export function getSocketPath(sessionId: string, config: MonocleConfig): string {
  return config.socket_path.replace('{{sessionId}}', sessionId);
}
