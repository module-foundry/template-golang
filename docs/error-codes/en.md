# Error codes

Public contract for clients (frontend). Codes are stable: never rename or
remove without a `BREAKING CHANGE`. Reuse a code before adding a new one;
specifics belong to `message`.

## API errors (4xx)

| Code | HTTP | When |
| --- | --- | --- |
| `VALIDATION_FAILED` | 422 | malformed/strict JSON or DTO validation failure |
| `UNAUTHORIZED` | 401 | missing or invalid token |
| `TOKEN_EXPIRED` | 401 | expired token (client should refresh) |
| `FORBIDDEN` | 403 | authenticated but not allowed |
| `NOT_FOUND` | 404 | entity does not exist |
| `CONFLICT` | 409 | state conflict (duplicate, version mismatch) |
| `RATE_LIMITED` | 429 | too many requests |
| `PAYLOAD_TOO_LARGE` | 413 | request body exceeds the limit |

## Runtime errors (5xx)

The client always receives a generic message; details stay in logs
(`type`, `message`, `trace`, `error`).

| Code | HTTP | When |
| --- | --- | --- |
| `INTERNAL` | 500 | unknown failure |
| `DB_QUERY_FAILED` | 500 | query failed |
| `DB_CONNECT_FAILED` | 500 | database unavailable |
| `DB_MIGRATION_FAILED` | 500 | startup migration failed |
| `REDIS_FAILED` | 500 | Redis unavailable |
| `HTTP_CLIENT_FAILED` | 500 | outbound HTTP call failed |
| `PANIC` | 500 | recovered panic |
| `CONFIG_INVALID` | 500 | invalid runtime configuration |

## Response shape

Errors use the `error` envelope:

```json
{
  "error": {
    "type": "VALIDATION_FAILED",
    "message": "email is invalid"
  }
}
```

Successful bodies (`result`) and pagination (`result.pagination`) follow
`docs/api-conventions/`.
