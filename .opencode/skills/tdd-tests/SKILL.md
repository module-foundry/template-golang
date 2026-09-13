---
name: tdd-tests
description: Use when writing or fixing tests, coverage or benchmarks in Russian or English ("тесты", "TDD", "покрытие", "бенчмарк", "tests", "coverage", "benchmark"). Stdlib testing only, red-green-refactor, contract tests for pkg/* with a 90% gate.
---

# TDD tests

## Workflow

1. Red: write a failing test describing the contract (handler/service behaviour,
   boundary, error path). Bug fixes start with a reproducing test.
2. Green: minimal implementation.
3. Refactor with tests green.

## Style

- Stdlib `testing` only; no testify/testcontainers/mockery.
- Table-driven with `t.Run`, `t.Helper()`, `t.Cleanup()`.
- Compare with `reflect.DeepEqual`/`slices.Equal`/`maps.Equal` or explicit
  checks.
- Fakes are hand-written small types implementing the consumer interface; keep
  them in the test file.
- Fiber handlers are tested through `app.Test(httptest.NewRequest(...))` with
  the project `ErrorHandler`; assert status codes and the error envelope.

## Coverage contract

- `pkg/*` every exported symbol has contract tests, gate >= 90%:
  `task test:cover`.
- Golden tests protect: `docs/openapi.json`, `docs/error-codes/{en,ru}.md`,
  error envelope/log field shapes.
- Exported API changes are breaking: update tests and
  `docs/architecture/{en,ru}.md` in the same change.

## Benchmarks

- `testing.B` with `b.ReportAllocs()`; run `task bench`.
- Add benchmarks for hot paths (JSON, auth, errors, repositories).
- Optimization without before/after numbers is rejected.

## Integration

- Tag `//go:build integration`; PostgreSQL/Redis from docker-compose or CI
  services via `TEST_POSTGRES_DSN`/`TEST_REDIS_ADDR`.
- Verify `db` tag mapping, auto-migrations and auth flow against real services.
