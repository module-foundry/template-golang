# API conventions

How response bodies are shaped so the API stays obvious and cheap to consume.
The rules are implemented in `pkg/response` and guarded by golden tests.

## Success

Every successful body wraps the payload once under `result`:

```json
{"result": {"token": "jwt", "user_id": "42"}}
```

- `result` holds exactly the endpoint DTO; access is O(1) (`body.result.field`).
- No extra nesting: `result.data`, `result.item.data` are forbidden.
- Collections are arrays under `result`; dictionary-shaped data stays an object
  (`{"x": 1}`), never an array of key/value pairs the client must fold.
- 204 responses have no body; do not send `{"result": null}`.
- HTTP status keeps its meaning: 200/201/204.

## Pagination

Page metadata lives at `result.pagination`, never at the envelope root:

```json
{
  "result": {
    "items": [{"id": "1"}],
    "pagination": {"page": 2, "per_page": 20, "total": 57}
  }
}
```

- `page`, `per_page`, `total` are the canonical fields
  (`pkg/response.Pagination`).
- `total_pages` is derived from `total` / `per_page`; it is not sent twice.
- One endpoint returns one collection; clients move between pages with query
  parameters, not by iterating or re-requesting per item.

## Errors

Errors never go inside `result`; they use the `error` envelope with a stable
`type`. See `docs/error-codes/`.

## Anti-patterns

- Envelope matryoshkas: `result -> data -> payload`.
- Arrays of one object "for uniformity" (forces `[0]` or a lookup).
- Shapes that make the client iterate to rebuild one entity (O(n) where O(1)
  is enough).
- Wrapper objects `{value, label, meta}` around a scalar without a reason.
- Different shapes per endpoint (`result` here, a bare object there).
- The same data both at the root and inside `result`.

## Rules

- The envelope is produced only by `pkg/response`; handlers and services never
  build it by hand.
- The OpenAPI schema must match the runtime body (`result` is required); the
  golden spec guards this.
- Extend payloads by adding fields; never rename or remove them.
