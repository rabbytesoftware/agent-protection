package manifest

const config = `
import tsParser from '@typescript-eslint/parser';
export default [
  {
    files: ['**/*.{ts,tsx}'],
    ignores: ['**/node_modules/**', '**/dist/**', '**/build/**', '**/*.config.*', '**/*.gen.ts', '**/*.test.*', '**/*.spec.*', '**/e2e/**', '**/__tests__/**', '**/__mocks__/**'],
    languageOptions: {
      parser: tsParser,
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    rules: {
      'complexity': ['warn', 8],
      'max-depth': ['warn', 2],
      'max-lines-per-function': ['warn', { max: 40, skipBlankLines: true, skipComments: true }],
      'max-params': ['warn', 4],
      'no-magic-numbers': ['warn', { ignore: [0, 1, -1, 200, 201, 204, 400, 401, 403, 404, 409, 422, 429, 500, 502, 503], ignoreArrayIndexes: true, enforceConst: true }],
      'no-else-return': ['warn', { allowElseIf: false }],
      'no-lonely-if': 'warn',
      'no-console': 'warn',
      'no-param-reassign': ['warn', { props: true }],
      'no-var': 'warn',
      'prefer-const': 'warn',
      'eqeqeq': ['warn', 'always'],
      'max-classes-per-file': ['warn', 1],
    },
  },
];
`
