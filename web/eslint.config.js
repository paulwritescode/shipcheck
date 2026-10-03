import js from '@eslint/js';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsparser from '@typescript-eslint/parser';
import reactHooks from 'eslint-plugin-react-hooks';

export default [
  { ignores: ['dist', 'node_modules'] },
  js.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: tsparser,
      parserOptions: { ecmaFeatures: { jsx: true } },
      globals: {
        // Browser globals
        document: 'readonly',
        window: 'readonly',
        fetch: 'readonly',
        console: 'readonly',
        navigator: 'readonly',
        // Vitest test globals (config has test.globals: true)
        describe: 'readonly',
        it: 'readonly',
        expect: 'readonly',
        vi: 'readonly',
        beforeEach: 'readonly',
        afterEach: 'readonly',
        // Node globals (used in vite.config.ts)
        __dirname: 'readonly',
        process: 'readonly',
        // React + DOM lib types used across the SPA
        React: 'readonly',
        HTMLInputElement: 'readonly',
        HTMLTextAreaElement: 'readonly',
        HTMLSelectElement: 'readonly',
        HTMLElement: 'readonly',
        HTMLButtonElement: 'readonly',
      },
    },
    plugins: {
      '@typescript-eslint': tseslint,
      'react-hooks': reactHooks,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      // TypeScript's own no-unused-vars is used; disable the base rule.
      'no-unused-vars': 'off',
    },
  },
];
