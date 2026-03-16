import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as os from 'node:os';
import { loadConfig, DEFAULT_CONFIG, getSocketPath } from '../../src/core/config.js';

vi.mock('node:fs');

const mockedFs = vi.mocked(fs);

describe('config', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    mockedFs.readFileSync.mockImplementation(() => {
      throw new Error('ENOENT');
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('loadConfig', () => {
    it('loads defaults when no config files exist', () => {
      const config = loadConfig();

      expect(config.mode).toBe('review');
      expect(config.gate_patterns).toEqual([]);
      expect(config.ignore_patterns).toEqual(DEFAULT_CONFIG.ignore_patterns);
      expect(config.socket_path).toContain('monocle-{{sessionId}}.sock');
    });

    it('merges global config over defaults', () => {
      const globalPath = path.join(
        process.env['XDG_CONFIG_HOME'] || path.join(os.homedir(), '.config'),
        'monocle',
        'config.json',
      );

      mockedFs.readFileSync.mockImplementation((filePath: fs.PathOrFileDescriptor) => {
        if (String(filePath) === globalPath) {
          return JSON.stringify({ mode: 'gate', gate_patterns: ['src/**'] });
        }
        throw new Error('ENOENT');
      });

      const config = loadConfig();

      expect(config.mode).toBe('gate');
      expect(config.gate_patterns).toEqual(['src/**']);
      // Defaults preserved for unset keys
      expect(config.ignore_patterns).toEqual(DEFAULT_CONFIG.ignore_patterns);
    });

    it('merges project config over global config', () => {
      const globalPath = path.join(
        process.env['XDG_CONFIG_HOME'] || path.join(os.homedir(), '.config'),
        'monocle',
        'config.json',
      );
      const projectRoot = '/tmp/my-project';
      const projectPath = path.join(projectRoot, '.monocle', 'config.json');

      mockedFs.readFileSync.mockImplementation((filePath: fs.PathOrFileDescriptor) => {
        if (String(filePath) === globalPath) {
          return JSON.stringify({ mode: 'gate', gate_patterns: ['src/**'] });
        }
        if (String(filePath) === projectPath) {
          return JSON.stringify({ mode: 'review', ignore_patterns: ['vendor/**'] });
        }
        throw new Error('ENOENT');
      });

      const config = loadConfig(projectRoot);

      // Project overrides global
      expect(config.mode).toBe('review');
      // Project override replaces the array
      expect(config.ignore_patterns).toEqual(['vendor/**']);
      // Global value preserved when not overridden by project
      expect(config.gate_patterns).toEqual(['src/**']);
    });

    it('handles malformed JSON gracefully', () => {
      const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      const globalPath = path.join(
        process.env['XDG_CONFIG_HOME'] || path.join(os.homedir(), '.config'),
        'monocle',
        'config.json',
      );

      mockedFs.readFileSync.mockImplementation((filePath: fs.PathOrFileDescriptor) => {
        if (String(filePath) === globalPath) {
          return '{ invalid json';
        }
        throw new Error('ENOENT');
      });

      const config = loadConfig();

      // Falls back to defaults
      expect(config.mode).toBe('review');
      expect(config.gate_patterns).toEqual([]);
      // No crash
      warnSpy.mockRestore();
    });

    it('warns on unknown config keys', () => {
      const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      const globalPath = path.join(
        process.env['XDG_CONFIG_HOME'] || path.join(os.homedir(), '.config'),
        'monocle',
        'config.json',
      );

      mockedFs.readFileSync.mockImplementation((filePath: fs.PathOrFileDescriptor) => {
        if (String(filePath) === globalPath) {
          return JSON.stringify({ mode: 'review', unknown_key: true });
        }
        throw new Error('ENOENT');
      });

      loadConfig();

      expect(warnSpy).toHaveBeenCalledWith(
        expect.stringContaining('unknown config key "unknown_key"'),
      );
      warnSpy.mockRestore();
    });

    it('warns on invalid mode value', () => {
      const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      const globalPath = path.join(
        process.env['XDG_CONFIG_HOME'] || path.join(os.homedir(), '.config'),
        'monocle',
        'config.json',
      );

      mockedFs.readFileSync.mockImplementation((filePath: fs.PathOrFileDescriptor) => {
        if (String(filePath) === globalPath) {
          return JSON.stringify({ mode: 'invalid' });
        }
        throw new Error('ENOENT');
      });

      const config = loadConfig();

      expect(warnSpy).toHaveBeenCalledWith(expect.stringContaining('invalid mode'));
      expect(config.mode).toBe('review'); // Falls back to default
      warnSpy.mockRestore();
    });
  });

  describe('getSocketPath', () => {
    it('resolves sessionId in socket path template', () => {
      const config = { ...DEFAULT_CONFIG };
      const result = getSocketPath('abc-123', config);

      expect(result).toContain('monocle-abc-123.sock');
      expect(result).not.toContain('{{sessionId}}');
    });

    it('works with custom socket_path', () => {
      const config = { ...DEFAULT_CONFIG, socket_path: '/run/monocle/{{sessionId}}.sock' };
      const result = getSocketPath('test-session', config);

      expect(result).toBe('/run/monocle/test-session.sock');
    });
  });
});
