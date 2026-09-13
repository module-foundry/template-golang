# Tests instructions

## TDD workflow (red-green-refactor)

1. Write a failing test that describes the contract or the bug.
2. Implement the minimum to make it pass.
3. Refactor with tests green.

Bug fixes always start with a test that reproduces the bug.

## Stack

- Stdlib `testing` only. No testify/testcontainers/mockery.
- Comparisons use `reflect.DeepEqual`, `slices.Equal`, `maps.Equal` or explicit
  checks.
- Mocks are hand-written fakes that implement the consumer interface; they live
  next to the test that uses them.
- Test files sit next to the code under test (`*_test.go`), table-driven with
  `t.Run` subtests, `t.Helper()` for helpers, `t.Cleanup()` for resources.

## Coverage contract

- Every exported symbol of `pkg/*` has contract tests, including boundaries and
  error paths. A change to exported API is a breaking change: tests and
  `docs/architecture/{en,ru}.md` are updated in the same change.
- `pkg/*` coverage gate is >= 90% and only grows (`task test:cover`).
- Golden tests protect formats: success envelope (`result`, `result.pagination`),
  error envelope, log fields, JWT claims, `docs/openapi.json`,
  `docs/error-codes/{en,ru}.md`.

## Fuzzing and benchmarks

- Fuzz parsers and token logic (`jwtx`, `validx`) with `testing.F`.
- Benchmarks for hot paths use `testing.B` (`B.Loop` where applicable) and must
  report allocations (`-benchmem`). Any performance change is accepted only with
  before/after numbers.

## Integration tests

- Build tag `integration` (`//go:build integration`).
- Real PostgreSQL/Redis come from docker-compose or CI services; DSNs are read
  from `TEST_POSTGRES_DSN` and `TEST_REDIS_ADDR`.
- Integration tests verify `db` tag mapping, auto-migrations, auth flow and
  repository contracts.
