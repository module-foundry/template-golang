# Architecture

## Overview

Single entry point: `cmd/api` (flags `-check-env`, `-dump-openapi`,
`-healthcheck`). All wiring lives in `internal/app`: config, logger, PostgreSQL,
Redis, migrations, Fiber, routes.

Layers:

```
router.go (module composition root: repo -> service -> handler, routes)
  -> handler (HTTP, typed via pkg/httpx)
    -> service (business rules, declares interfaces)
      -> repository (pgx, SQL only)
```

Dependencies point down only. The service never imports Fiber; the repository
never knows about HTTP. Each module is a vertical slice
(`internal/modules/<name>/`) with its own `router.go`; `internal/app/routes.go`
only calls `Register` on the public or protected router.

## Packages

| Package | Responsibility |
| --- | --- |
| `pkg/httpx` | typed route helpers + route registry for OpenAPI |
| `pkg/openapi` | runtime OpenAPI 3.0.3 + Scalar UI (embedded assets) |
| `pkg/apperror` | error registry, constructors, single log/response boundary |
| `pkg/logger` | slog setup, request-scoped logger, reporter seam |
| `pkg/jsonx` | `encoding/json/v2`: marshal/unmarshal, strict DTO parsing |
| `pkg/jwtx` | JWT sign/parse, token extraction (cookie, then header) |
| `pkg/postgres` | pgxpool init, healthcheck, `WithTx` |
| `pkg/redisx` | Redis client init, healthcheck |
| `pkg/validx` | shared validator, maps failures to `VALIDATION_FAILED` |
| `pkg/response` | success response helpers |

## Request flow

1. `RequestID` middleware assigns/propagates `X-Request-ID`.
2. `RequestLogger` builds a request-scoped logger and puts it in the context.
3. `recover`, `cors`, `helmet` (and `pprof` when enabled).
4. Routes are mounted under `API_BASE_PATH` (default `/api/v1`). `/auth/*` routes
   and docs are public; everything that needs auth is mounted on the protected
   group (`middleware.Auth`): cookie `access_token` first, then
   `Authorization: Bearer`. In Fiber the group middleware only guards routes
   registered after it, so public routes and docs are registered first.
5. Typed handler: strict JSON parse -> validation -> service.
6. Errors are returned and formatted/logged once by `apperror.Handler`.

## Errors

Two kinds: `API` (4xx, safe message) and `Runtime` (5xx, generic message,
cause only in logs). Codes are constants in `pkg/apperror/codes.go` with a
minimal, stable set; HTTP statuses come from `fiber.Status*` constants.
Success bodies go through `pkg/response` and wrap the payload as
`{"result": ...}`; errors use the `error` envelope, so clients can tell the two
apart without inspecting the status code.

Error log fields: `type`, `message`, `trace`, `kind`, `http_status`,
`request_id`, `user_id`, and `error` (cause) for runtime errors.

## Configuration

`.env` for local development (real env vars win), `.env.example` always in sync.
`internal/config.Load()` decodes and validates everything, including production
invariants (strong JWT secret, `COOKIE_SECURE=true`, no wildcard CORS).
Comma-separated values are parsed into typed slices.

## Migrations

`migrations/*.sql` (goose Up/Down) are embedded and applied at startup when
`DB_AUTO_MIGRATE=true`, guarded by a PostgreSQL session advisory lock. Manual
`task migrate:*` commands remain for rollback and development.

## API contract

Routes are registered only through `pkg/httpx`; the registry feeds runtime
OpenAPI (`{API_BASE_PATH}/openapi.json`, `{API_BASE_PATH}/docs`). `docs/openapi.json` is golden-tested, route
tables are snapshot-tested, and `oasdiff breaking` runs in CI. Changing the
public contract without an explicit request fails the pipeline.

## Documentation

`docs/<topic>/{en,ru}.md`; ADRs in `docs/adr/NNNN-<slug>/{en,ru}.md`.
`README.md` is English-only and short.
