# Skills

AI skills live in `.opencode/skills/<name>/SKILL.md` and are loaded on demand.
Routing lives in `AGENTS.md`; hard rules live in `.opencode/instructions/`.

| Skill | Trigger (RU/EN) | Purpose |
| --- | --- | --- |
| `add-module` | добавь endpoint/фичу/модуль; add feature | TDD-first module: handler/service/repository/dto + route registration |
| `pgx-queries` | запрос к БД; repository; SQL | `db` tags + `RowToStructByName`, transactions, query rules |
| `create-migration` | миграция; migration | goose naming, up/down, auto-start migrations, expand/contract |
| `error-handling` | ошибка; error code; лог | registry, kinds, fiber statuses, 120-char message, trace flag |
| `tdd-tests` | тесты; TDD; покрытие; benchmark | red-green-refactor, stdlib testing, contract tests, coverage gate |
| `add-realtime` | realtime; websocket; ws | creates the isolated WS layer when explicitly requested |
| `openapi` | swagger; openapi; orval | typed routes, runtime spec, Scalar UI, golden workflow |
| `commit` | закоммить; commit | Conventional Commits, Go commitlint; only on explicit request |
| `docs` | документация; docs; readme; adr | `docs/<topic>/{en,ru}.md`, EN-only README, ADRs |
| `preflight` | готово; проверь; done | definition of done: `task verify`, `task api:check`, docs sync |
| `add-skill` | добавь скилл; add skill | SKILL.md format, registration, routing table, docs sync |

## How selection works

- Each `description` contains literal trigger words; the model sees them and
  loads only relevant skills.
- `AGENTS.md` defines the mandatory workflow: classify -> load skill -> execute
  -> `preflight`.
- Chains: `add-module` -> `tdd-tests` + `pgx-queries`; `create-migration` ->
  `pgx-queries`; API changes -> `openapi`.
- After changing skills, use `add-skill` to keep this table and `AGENTS.md` in
  sync.
