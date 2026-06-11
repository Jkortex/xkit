import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    projects: [
      {
        test: {
          include: ['packages/*/tests/**/*.test.ts'],
          environment: 'node',
          name: 'node',
        },
      },
      {
        test: {
          include: ['packages/hotkeys/src/**/*.test.ts'],
          environment: 'happy-dom',
          name: 'hotkeys',
        },
      },
    ],
  },
});
