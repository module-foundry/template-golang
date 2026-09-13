# AGENTS.md

Операционный хаб проекта. Держи файл коротким: детали — в instructions и скиллах.

## Алгоритм работы

1. Классифицируй задачу по таблице маршрутизации и **загрузи скилл через tool `skill` до написания кода**.
2. Работай строго по шагам скилла (он самодостаточен: команды, настройки, шаблоны).
3. Перед словом «готово» обязательно выполни скилл `preflight` (`task verify`, `task api:check`).
4. Не меняй публичный контракт API (роуты, DTO, коды ошибок, статусы) без явного запроса.

## Маршрутизация задач

| Задача / триггер | Скилл |
| --- | --- |
| «добавь endpoint/фичу/модуль» | `add-module` (цепочка: `tdd-tests`, `pgx-queries`) |
| «запрос к БД», «repository-метод» | `pgx-queries` |
| «миграция», «изменить схему» | `create-migration` (`pgx-queries`) |
| «ошибка», «лог», «обработка ошибок» | `error-handling` |
| «тесты», «TDD», «покрытие», «бенчмарк» | `tdd-tests` |
| «realtime», «websocket», «ws» | `add-realtime` |
| «swagger», «openapi», «orval» | `openapi` |
| «закоммить», «commit» | `commit` (только явный запрос или явная точка коммита) |
| «документация», «docs», «readme», «adr» | `docs` |
| «готово», «проверь», «перед продом» | `preflight` |
| «добавь скилл», «обнови правила AI» | `add-skill` |

## Команды

```bash
task setup      # инструменты + зависимости
task dev        # hot reload (Air); миграции применит старт приложения
task test       # unit-тесты
task test:cover # покрытие + gate pkg/* >= 90%
task bench      # бенчмарки hot paths
task verify     # полный локальный гейт
task api:check  # контракт API: snapshot роутов + golden OpenAPI
task env:check  # проверка .env
```

F5 в VS Code запускает API под dlv (профиль «API (debug)»).

## Карта проекта

- `cmd/api` — единственная точка входа (`-check-env`, `-dump-openapi`, `-healthcheck`).
- `internal/app` — сборка зависимостей, роуты, lifecycle.
- `internal/config` — типизированный конфиг и валидация.
- `internal/middleware` — auth (JWT), request id/log.
- `internal/modules/<name>` — вертикальный срез: `router.go` (composition root) + handler + service + repository + dto.
- `pkg/*` — переиспользуемые пакеты по зонам ответственности (без локальных utils).
- `migrations` — SQL + автозапуск goose.
- `docs/<topic>/{en,ru}.md` — документация (README только EN).
- `.opencode/` — правила (`instructions`) и скиллы.

## Правила верхнего уровня

- `.opencode/instructions/*` загружаются всегда и обязательны к соблюдению.
- Публичный API заморожен: `task api:check` должен быть зелёным.
- Комментарии и код — на английском; доки — `docs/<topic>/{en,ru}.md`.
- Новые зависимости, коды ошибок и env-переменные — только с обоснованием и обновлением docs/.env.example.
