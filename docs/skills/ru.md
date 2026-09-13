# Скиллы

AI-скиллы лежат в `.opencode/skills/<name>/SKILL.md` и загружаются по
требованию. Маршрутизация — в `AGENTS.md`, жёсткие правила — в
`.opencode/instructions/`.

| Скилл | Триггер (RU/EN) | Назначение |
| --- | --- | --- |
| `add-module` | добавь endpoint/фичу/модуль; add feature | TDD-модуль: handler/service/repository/dto + регистрация роута |
| `pgx-queries` | запрос к БД; repository; SQL | `db`-теги + `RowToStructByName`, транзакции, правила запросов |
| `create-migration` | миграция; migration | goose naming, up/down, автозапуск, expand/contract |
| `error-handling` | ошибка; error code; лог | реестр, kind, fiber-статусы, лимит 120, trace-флаг |
| `tdd-tests` | тесты; TDD; покрытие; benchmark | red-green-refactor, stdlib, контрактные тесты, gate |
| `add-realtime` | realtime; websocket; ws | создаёт изолированный WS-слой по явному запросу |
| `openapi` | swagger; openapi; orval | типизированные роуты, рантайм-спека, Scalar, golden |
| `commit` | закоммить; commit | Conventional Commits, Go-commitlint; только по явному запросу |
| `docs` | документация; docs; readme; adr | `docs/<topic>/{en,ru}.md`, README только EN, ADR |
| `preflight` | готово; проверь; done | definition of done: `task verify`, `task api:check`, синк доков |
| `add-skill` | добавь скилл; add skill | формат SKILL.md, регистрация, маршрутизация, синк доков |

## Как работает автовыбор

- В `description` есть триггерные слова; модель видит их и загружает только
  нужные скиллы.
- `AGENTS.md` задаёт обязательный цикл: классифицировать -> загрузить скилл ->
  выполнить -> `preflight`.
- Цепочки: `add-module` -> `tdd-tests` + `pgx-queries`; `create-migration` ->
  `pgx-queries`; изменения API -> `openapi`.
- После изменения скиллов используйте `add-skill`, чтобы таблица здесь и в
  `AGENTS.md` оставалась актуальной.
