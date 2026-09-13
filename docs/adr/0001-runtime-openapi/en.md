# ADR 0001: Runtime OpenAPI without annotations

## Context

Orval needs an OpenAPI document for the frontend. The common approach (swaggo
annotations in handlers) requires doc comments on every endpoint, which turns
handlers into noise and drifts from the code.

## Decision

- Register every public route through `pkg/httpx` typed helpers
  (`Get/Post/Patch/Delete` with `Req`/`Resp` type parameters).
- `pkg/httpx` records method, path, operationId, DTO types and auth flag.
- `pkg/openapi` builds an OpenAPI 3.0.3 document at runtime via reflection over
  the recorded DTO types; Scalar UI assets are embedded.
- The document is served at `/openapi.json` and `/docs`; `docs/openapi.json` is
  a golden file guarded by a test; CI runs `oasdiff breaking`.

OpenAPI 3.0.3 (not 3.1) is chosen for maximum Orval compatibility.

## Consequences

- Handlers contain no doc comments and no annotations.
- Adding an endpoint automatically updates the spec after restart.
- The registry is also the source for route snapshot tests, which blocks
  accidental API changes.
- Exotic types (polymorphism, custom serializers) require extending
  `pkg/openapi` explicitly instead of annotating handlers.
