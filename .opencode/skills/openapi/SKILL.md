---
name: openapi
description: Use when working with Swagger/OpenAPI, the API reference UI or Orval generation in Russian or English ("swagger", "openapi", "orval", "схема api", "/docs", "спека"). Explains runtime generation without annotations and the golden contract workflow.
---

# OpenAPI / Swagger

## How it works

- Routes are registered only through `pkg/httpx` typed helpers
  (`httpx.Get/Post/Patch/Delete`, `Protected()`, `Tag("module")`). The registry
  records method, path, operationId (from the handler name), request/response
  types and auth flag.
- `pkg/openapi` builds an OpenAPI 3.0.3 document at runtime from the registry
  and reflection over DTOs. No comments or annotations in handlers.
- Endpoints (enabled by `FEATURE_OPENAPI_ENABLED=true`), mounted under
  `API_BASE_PATH` (default `/api/v1`):
  - `GET {API_BASE_PATH}/openapi.json` — the document;
  - `GET {API_BASE_PATH}/docs` — Scalar UI (assets embedded, no CDN, no build step).
- Orval consumes `http://localhost:3001/api/v1/openapi.json` in dev or the golden
  `docs/openapi.json` in CI.

## DTO requirements

- `json` tags for names; pointers or `omitempty` mark optional fields;
- `validate:"required"` maps to required properties;
- `enum:"a,b,c"`, `format:"uuid|date-time|..."`, `example:"..."` enrich schema.

## Contract workflow

1. Add/change routes through `httpx` helpers.
2. Regenerate the golden file: `task openapi:dump`.
3. Run `task api:check` (route snapshot + golden OpenAPI tests).
4. A breaking change (removed path, renamed field, new required field) is
   allowed only on an explicit request: update the golden file and mark the
   commit `BREAKING CHANGE`.
5. In CI `oasdiff breaking <base> docs/openapi.json` guards against accidental
   regressions.

## Rules

- Never write OpenAPI by hand and never add handler comments for docs.
- If a type cannot be expressed by reflection, extend `pkg/openapi` with a small
  explicit override instead of inventing annotations.
