import js from '@eslint/js';
import tseslint from '@typescript-eslint/eslint-plugin';
import parser from '@typescript-eslint/parser';

const globals = {
  browser: true,
  console: true,
  setInterval: true,
  clearInterval: true,
  setTimeout: true,
  clearTimeout: true,
  Window: true,
  window: true,
  Blob: true,
  URL: true,
  KeyboardEvent: true,
  document: true,
  HTMLDivElement: true,
  HTMLSelectElement: true,
  HTMLInputElement: true,
  HTMLButtonElement: true,
  global: true,
  React: true
};

export default [
  {
    ignores: ['dist', 'node_modules', 'wailsjs']
  },
  js.configs.recommended,
  {
    files: ['src/**/*.{js,jsx,ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2021,
      sourceType: 'module',
      parser,
      parserOptions: {
        ecmaFeatures: {
          jsx: true
        }
      },
      globals
    },
    plugins: {
      '@typescript-eslint': tseslint
    },
    rules: {
      ...tseslint.configs.recommended.rules,
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'warn',
      'no-console': 'off'
    }
  }
];
