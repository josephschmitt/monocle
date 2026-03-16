import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    root: '.',
    passWithNoTests: true,
  },
  resolve: {
    alias: {
      '@core': './src/core',
      '@adapters': './src/adapters',
      '@db': './src/db',
      '@types': './src/types',
      '@tui': './src/tui',
      '@hooks': './src/hooks',
    },
  },
});
