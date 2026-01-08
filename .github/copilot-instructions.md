# Copilot Instructions

## Principles
- Default to existing i18n keys; add locale entries when needed instead of hardcoding strings.
- Match lint/format rules (Go fmt/go vet, frontend ESLint/Prettier); keep changes minimal and consistent with nearby code.
- Preserve or improve test coverage; add focused tests for new logic or fixes.

## Required checks before handoff
- Go: `make test`
- Frontend: `make frontend-test` or `npm test`
- Full suite: `make test-all`
- Lint: `make lint` (Go + TypeScript)

## Code organization
- Go: Package-level structs with clear fields; exported types in PascalCase, unexported in camelCase.
- TypeScript: Use interfaces for data models (e.g., `dtoSong`, `SelectedSong`); follow existing dto prefix patterns.
- Components: Place React components in `frontend/src/components/<ComponentName>/` with `index.tsx`, `index.module.less`, and `tests/<ComponentName>.test.tsx`.
- Tests: Go tests as `*_test.go` beside source; fixtures in `testdata/`. Frontend tests in `tests/` inside each component folder.

## Testing standards
- Go: Table-driven tests with descriptive names; use setup/teardown helpers (e.g., `setupTestDB`, `teardownTestDB`).
- Frontend: Vitest + `@testing-library/react`; wrap components with required providers/contexts; cover interactions and edge cases.
- Mock external dependencies; avoid network/file I/O in unit tests (use temp files or in-memory data).

## Style conventions
- Go: Run `go fmt`; add brief comments on exported items (e.g., `// App struct`).
- TypeScript: Follow ESLint rules; avoid unused variables; prefer `const`/`let`.
- CSS: Use LESS modules with semantic class names; import shared vars from `vars.less`.
- Naming: Prefer descriptive names (e.g., `GetSongAuthors`, `renderWithContext`); abbreviate only when conventional (e.g., `ctx`, `err`).
