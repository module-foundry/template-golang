# AGENTS.md

Project operations hub. Keep this file short: details belong in instructions and skills.

## Workflow

1. Classify the task using the routing table and **load the skill through the `skill` tool before writing code**.
2. Follow the skill steps exactly (it is self-contained: commands, settings, templates).
3. Before saying "done", always run the `preflight` skill (`task verify`, `task api:check`).
4. Do not change the public API contract (routes, DTOs, error codes, statuses) without an explicit request.

## Task Routing

| Task / trigger | Skill |
| --- | --- |
| "add an endpoint/feature/module" | `add-module` (chain: `tdd-tests`, `pgx-queries`) |
| "database query", "repository method" | `pgx-queries` |
| "migration", "change the schema" | `create-migration` (`pgx-queries`) |
| "error", "log", "error handling" | `error-handling` |
| "tests", "TDD", "coverage", "benchmark" | `tdd-tests` |
| «realtime», «websocket», «ws» | `add-realtime` |
| «swagger», «openapi», «orval» | `openapi` |
| "commit", "make a commit" | `commit` (explicit request or explicit commit point only) |
| "documentation", "docs", "readme", "adr" | `docs` |
| "done", "check", "before production" | `preflight` |
| "add a skill", "update AI rules" | `add-skill` |

## Commands

```bash
task setup      # tools and dependencies
task dev        # hot reload (Air); application startup applies migrations
task test       # unit tests
task test:cover # coverage and pkg/* >= 90% gate
task bench      # hot-path benchmarks
task verify     # full local gate
task api:check  # API contract: route snapshot and golden OpenAPI
task env:check  # validate .env
```

F5 in VS Code starts the API under dlv (the "API (debug)" profile).

## Project Map

- `cmd/api` - the only entry point (`-check-env`, `-dump-openapi`, `-healthcheck`).
- `internal/app` - dependency wiring, routes, lifecycle.
- `internal/config` - typed configuration and validation.
- `internal/middleware` - auth (JWT), request ID/logging.
- `internal/modules/<name>` - vertical slice: `router.go` (composition root) + handler + service + repository + DTO.
- `pkg/*` - reusable packages by responsibility area (no local utils).
- `migrations` - SQL and automatic goose execution.
- `docs/<topic>/{en,ru}.md` - documentation (README in English only).
- `.opencode/` - rules (`instructions`) and skills.

## Top-Level Rules

- `.opencode/instructions/*` are always loaded and must be followed.
- The public API is frozen: `task api:check` must pass.
- Comments and code must be in English; documentation belongs in `docs/<topic>/{en,ru}.md`.
- Add dependencies, error codes, and environment variables only with justification and updates to docs/.env.example.
