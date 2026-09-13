---
name: error-handling
description: Use when adding or changing errors, error codes, logging of failures or panic handling in Russian or English ("ошибка", "лог", "обработка ошибок", "error", "error code", "logging", "panic"). Enforces the apperror registry, fiber statuses and minimal stable codes.
---

# Error handling

## Code registry

- Codes live only in `pkg/apperror/codes.go`; HTTP statuses use `fiber.Status*`
  constants (numeric literals are test-forbidden).
- Reuse first: one code covers a class of situations; specifics go to
  `message`/details. Add a new code only when the client must branch on it.
- Renaming or removing a code is a breaking change: update
  `docs/error-codes/{en,ru}.md` (golden-tested) and mark the commit
  `BREAKING CHANGE`.

## Kinds

| Kind | Use for | Status | Client message | Log |
| --- | --- | --- | --- | --- |
| `API` | client mistakes, expected business rejections | 4xx | the message itself | warn |
| `Runtime` | infrastructure/unknown failures | 5xx | generic (`internal server error`) | error + cause |

## Constructors

```go
apperror.API(apperror.CodeNotFound, "order not found")
apperror.Runtime(apperror.CodeDBQueryFailed, err)
apperror.Wrap(err, apperror.CodeInternal)
```

- `message` max 120 chars; the constructor truncates.
- `trace` is captured when `ERROR_TRACE_ENABLED=1`; full stack for 5xx only when
  `ERROR_STACKTRACE_ENABLED=1`.
- Log exactly once: return the error from handlers, `apperror.Handler` logs and
  formats the envelope. Never `log + return` in a layer.
- Panics are converted to `PANIC` by `recover` middleware.

## Response shape

```json
{"error": {"type": "NOT_FOUND", "message": "order not found"}}
```

Never leak causes, SQL or stack traces to clients.
