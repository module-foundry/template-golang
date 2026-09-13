---
name: add-module
description: Use when adding a new endpoint, feature or domain module in Russian or English ("добавь endpoint", "добавь фичу", "новый модуль", "add feature", "add endpoint", "new module"). TDD-first vertical slice with router.go as the module composition root.
---

# Add module (vertical slice)

## Layout

```
internal/modules/<name>/
├── router.go      # composition root: repo -> service -> handler + routes
├── handler.go     # HTTP only
├── service.go     # business rules, declares Repository interface
├── repository.go  # pgx only
└── dto.go         # request/response structs
```

## Steps

1. Scaffold: `task gen:module -- <name>` creates the full slice (router
   included).
2. **Tests first** (`tdd-tests` skill): handler/service tests; load
   `pgx-queries` if a repository is involved.
3. DTO (`dto.go`): `json` + `validate` tags, optional fields are pointers,
   `example` tags for OpenAPI.
4. Repository (`repository.go`): pgx, `db` tags + `RowToStructByName`,
   parameterized queries, `context.Context` first.
5. Service (`service.go`): declare the repository interface here (consumer
   side); return `*apperror.Error` for expected failures.
6. Handler (`handler.go`): typed methods
   `func (h *Handler) Method(ctx context.Context, req Req) (Resp, error)`
   (or a `HandlerWithCtx` variant if cookies/headers are needed). No SQL, no
   Fiber in the service.
7. **Router (`router.go`)**: declare `Deps` with only the infrastructure the
   module needs, build `NewXRepository -> NewService -> NewHandler`, and
   register typed routes via `pkg/httpx` (`httpx.Protected()` for
   authenticated ones, `httpx.Tag("<name>")`).
8. Wire in `internal/app/routes.go` — exactly one call:
   `mymodule.Register(protected, deps.Registry, mymodule.Deps{Pool: deps.Pool})`.
   Public modules use `public`, protected modules use `protected`.
9. Tests: `task test`; then `task api:check` (regenerate the golden spec with
   `task openapi:dump` if the route is new).
10. Finish with `preflight`.

## Rules

- Do not touch other modules, `pkg/*` or unrelated routes.
- Public contract (path, method, DTO fields, error codes) changes only on an
  explicit request.
- No comments-as-docs in handlers: OpenAPI comes from types and the registry.

## Reference

Existing example: `internal/modules/auth` (stub token endpoint with its own
`router.go`).
