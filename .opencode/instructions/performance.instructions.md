# Performance and allocations instructions

## Focus

Allocation awareness is part of the definition of done for hot paths:
middlewares (auth, request logging), JSON handling, repositories and future
realtime broadcasts.

## Required techniques

- Preallocate: `make([]T, 0, n)`, `slices.Grow`, `make(map[K]V, n)` when the
  size is known.
- Strings: `strings.Builder` with `Grow`, `strconv.Append*` instead of
  `fmt.Sprintf` in hot paths.
- Minimize `string <-> []byte` conversions; prefer `[]byte` APIs when
  appropriate. `unsafe` is forbidden.
- `sync.Pool` for buffers and heavy temporary structures (JSON, WS frames);
  always reset before returning an object to the pool.
- Small structs pass by value; large or mutated structs by pointer.
- `slog`: build attributes only when the level is enabled; reuse static
  `slog.Attr` values.
- Fiber: `c.Body()`/`c.Query()` may reference internal buffers — do not retain
  them after the handler returns (`Immutable=false`); copy deliberately.
- pgx: `RowToStructByName` without intermediate `map[string]any`; prepared
  statements; reuse structures when iterating batches.
- `encoding/json/v2`: use streaming (`UnmarshalRead`, `MarshalWrite`,
  `jsontext`) for large payloads instead of intermediate `[]byte`.

## Measurement

- `go test -run '^$' -bench . -benchmem ./pkg/...` (task `bench`).
- `testing.B.Loop` in new benchmarks; pprof under `FEATURE_PPROF_ENABLED`.
- Optimize only after a profile shows the problem. A performance change without
  before/after numbers is rejected.
- Keep benchmarks deterministic and free of external infrastructure.
