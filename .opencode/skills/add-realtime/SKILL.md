---
name: add-realtime
description: Use when adding realtime / WebSocket support in Russian or English ("добавь realtime", "нужен websocket", "ws-слой", "add websocket", "realtime"). Creates the isolated WS layer by project conventions; realtime is absent by default.
---

# Add realtime (WebSocket)

The template ships without WebSocket code on purpose. When realtime is
requested, create it in this order.

## Checklist

1. Dependency: `go get github.com/gofiber/contrib/v3/websocket`.
2. Package `internal/realtime`:
   - `hub.go`: `Hub` with `Register/Unregister/Broadcast`, buffered channels,
     non-blocking send (`select` + `default`).
   - `noop.go` is not needed; instead the route is mounted only when the flag is
     on.
   - `Close()` exists and is called during graceful shutdown.
3. Interface first: the consuming service declares
   `type Broadcaster interface { Broadcast(event Event) }`; `app.go` binds it to
   the hub. Business code imports no websocket library.
4. Config: add `FEATURE_WS_ENABLED=false` to `.env.example` and
   `internal/config` with validation.
5. Route: register `/ws` in `internal/app/routes.go` only when enabled.
   Auth uses the same `jwtx`: cookie or `?token=` (browsers cannot set
   WebSocket headers); unauthorized upgrade -> `401`.
6. JSON: typed event structs, serialize with `pkg/jsonx`; for large streams use
   `jsontext` streaming.
7. Tests (`tdd-tests`): hub with fake connections, ping/pong, auth rejection.
8. Docs: update `docs/architecture/{en,ru}.md` (both languages) and this skill
   if conventions changed.

## Invariants

- Publication never blocks an HTTP request.
- Payloads are encoded once per broadcast, not once per client.
- Every goroutine has an owner and an exit path.
- `go test -race ./internal/realtime/...` passes.
