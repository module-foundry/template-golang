# Project instructions

Always-on rules for this repository. No procedures or templates here — those
live in `.opencode/skills/`.

## Architecture

- Layers: `handler -> service -> repository`, dependencies point down only.
  `handler` never contains SQL; `service` never imports Fiber; `repository`
  never knows about HTTP.
- Interfaces are declared by the consumer (in `service.go`), implementations
  live next to them (`repository.go`).
- `internal/` — application code; `pkg/*` — reusable primitives grouped by a
  single responsibility (`jwtx`, `jsonx`, `validx`, `apperror`, ...).
- No local utilities: `utils.go`, `helpers.go`, `common.go`, `misc.go` are
  forbidden. A reusable primitive becomes a `pkg/` package named after its
  responsibility; package with two unrelated topics is split.
- No global state and no `init()` magic: everything is wired explicitly in
  `internal/app`.
- Adding a feature = `internal/modules/<name>/` (vertical slice: `router.go`,
  `handler.go`, `service.go`, `repository.go`, `dto.go`) + one `Register` call
  in `internal/app/routes.go`.
- `router.go` is the module composition root: it builds
  `repository -> service -> handler`, receives only the infrastructure it needs
  via `Deps`, and registers its own routes through `pkg/httpx`. The app does not
  know module internals.
- Module = one domain, 4–6 files. Shared code is extracted, never copied.

## Errors

- Every error crossing a boundary is `*apperror.Error`.
- `type` is a constant from `pkg/apperror/codes.go`; string literals are
  forbidden. HTTP statuses come from `fiber.Status*`/`http.Status*` constants;
  numeric literals are rejected by a test.
- Codes are minimal and stable: reuse an existing code before adding a new one.
  A new code is justified only when the client must branch on it and no existing
  code expresses the case. Renaming/removing a code is a breaking change.
- Kinds: `API` (4xx, safe message) and `Runtime` (5xx, generic message to the
  client, cause stays in logs).
- `message` is at most 120 characters; the constructor truncates and a test
  guards the limit.
- Log an error exactly once, at the HTTP boundary (`apperror.Handler`), never in
  every layer. Do not log and return the same error from a domain layer.

## Logging

- Use `log/slog` through `pkg/logger`; in layers use `logger.FromCtx(ctx)`.
- Structured fields only. Secrets and PII (Authorization, cookies, DSN,
  passwords, request bodies) never reach logs.
- Access logs: 5xx -> error, 4xx -> warn, otherwise info.

## JSON

- `encoding/json/v2` only, through `pkg/jsonx`. `encoding/json` v1 and
  third-party JSON libraries are forbidden.
- Inbound DTOs are parsed strictly: `jsonx.UnmarshalStrict` (unknown members are
  rejected) and then validated with `validx`. Unknown field -> `VALIDATION_FAILED`.
- `db` and `json` tags are never mixed on one struct: models are for the DB,
  DTOs are for the API.

## Database

- pgx/v5 only. Every row mapping uses `db:"col"` tags with
  `pgx.CollectRows(rows, pgx.RowToStructByName[T])` (or `...Lax` for optional
  columns). Manual `rows.Scan` and ORMs are forbidden.
- Parameterized queries only, `context.Context` is always the first argument.
- Transactions go through `postgres.WithTx`.

## Configuration

- Everything goes through `internal/config`; a new environment variable requires
  a typed field, validation, and an update to `.env.example` in the same change.
- Comma-separated lists are parsed in config into typed slices; `strings.Split`
  in business code is forbidden.

## HTTP and API contract

- All responses go through `pkg/response`; routes are registered only via
  `pkg/httpx` typed helpers (they feed the OpenAPI registry).
- The public contract (routes, paths, methods, `operationId`, DTO fields, error
  codes, statuses) is frozen. Change it only on an explicit request for that
  specific route/DTO; "improving" unrelated endpoints is forbidden.
- Run `task api:check` before saying a task is done; a conscious breaking change
  updates `docs/openapi.json` and is marked `BREAKING CHANGE` in the commit.

## AI workflow

- Classify the task with the routing table in `AGENTS.md` and load the relevant
  skill(s) before writing code.
- After changing skills, sync `docs/skills/{en,ru}.md` and the routing table.
- Commits follow Conventional Commits (`commitlint`) and are created only on an
  explicit request or an explicit commit point.

## Principles

### DRY — Don't Repeat Yourself

- One fact lives in one place: schema in migrations, SQL in repositories, error
  codes in the registry, env in `internal/config`, routes in `routes.go`.
- Knowledge is duplicated, not strings: if a rule must change in two places, it
  is a bug waiting to happen.
- Rule of three: extract to `pkg/` after the third repetition or as soon as a
  second module needs it. Speculative abstraction is forbidden.
- Review question: "if X changes, how many files do we edit?" More than one
  means the knowledge is not centralized.

### KISS — Keep It Simple, Stupid

- Simplest working solution: stdlib before dependencies, explicit code before
  magic, linear code before clever abstractions.
- Early returns, happy path left-aligned, minimal nesting.
- Nothing "for the future": an interface appears when there is a second
  implementation or a test fake, not earlier.
- DI by hand in `internal/app`, no config layering, no routing meta-programming.
- Exception: hot paths may be non-obvious, but they must stay local, isolated in
  `pkg/*`, and covered by a benchmark.

### SOLID in this project

- **S — Single Responsibility**: `handler` is HTTP only, `service` is business
  rules only, `repository` is SQL only; a `pkg/*` package covers one topic; a
  file has one reason to change. Check: "how many reasons force this type to
  change?" More than one — split.
- **O — Open/Closed**: extend by adding a module/implementation, not by editing
  foreign switches. Examples: a new error code is a registry entry, call sites
  stay untouched; realtime is added as a separate package by `add-realtime`;
  a new endpoint is a new module plus one line in `routes.go`.
- **L — Liskov Substitution**: the interface declared by the consumer is a
  contract; any implementation (pgx, fake, redis) behaves identically — same
  errors, same ordering, no panics or hidden side effects. Contract tests run
  against every implementation.
- **I — Interface Segregation**: interfaces are small (1–3 methods) and describe
  the consumer's need (`TokenParser`, `Broadcaster`). God interfaces are
  forbidden; interfaces are declared by the consumer.
- **D — Dependency Inversion**: services depend on interfaces, not on pgx/Redis;
  the module composition root (`router.go`) binds abstractions to concrete
  implementations, and `internal/app` only hands over infrastructure (`Deps`).
  Modules stay unit-testable without infrastructure.

### Inheritance and polymorphism in Go

- There is no classical inheritance; embedding and composition fill that role:
  `type Service struct { repo Repository; log *slog.Logger }`. Embedding is for
  reusing behavior, not for building type hierarchies.
- Polymorphism is interface-based: one service call, many implementations
  (Postgres/Redis/fake), selected in `router.go` of the module or `internal/app`
  by config or feature flag.
- Accept interfaces, return concrete types.
- A type switch is allowed only inside `pkg/apperror` (mapping to HTTP/log);
  business code never branches on concrete types.
- Composition over "base classes": small types + embedding + interfaces;
  `BaseService`-style patterns are forbidden.
