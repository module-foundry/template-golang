# Template Golang

Fast and safe Go backend template: Fiber v3, PostgreSQL (pgx), Redis, JWT auth, runtime OpenAPI (Scalar UI), embedded goose migrations and an AI-ready layer (instructions + skills).

## Requirements

- Go 1.27+
- Docker (PostgreSQL, Redis, optional container builds)
- [go-task](https://taskfile.dev) and [delve](https://github.com/go-delve/delve) (installed by `task setup`)
- VS Code with the `golang.go` extension for the shared debug profile

## Quick start

```bash
cp .env.example .env
task setup        # installs dev tools (dlv, goose, commitlint, oasdiff, air)
task dev          # hot reload (Air); migrations are applied on startup
```

VS Code: press `F5` to run the API under dlv (uses `.env`).

API prefix is `API_BASE_PATH` (default `/api/v1`).
Useful endpoints: `/healthz`, `/readyz`, `/api/v1/docs` (Scalar UI), `/api/v1/openapi.json`.
Auth stub: `POST /api/v1/auth/mini-apps/telegram`.

## Documentation

Detailed docs live in [`docs/`](docs/): architecture, skills, error codes and ADRs — each topic in `en.md` and `ru.md`.

## License

MIT — see [LICENSE](LICENSE).
