# Архитектура

## Обзор

Единственная точка входа: `cmd/api` (флаги `-check-env`, `-dump-openapi`,
`-healthcheck`). Вся сборка зависимостей — в `internal/app`: конфиг, логгер,
PostgreSQL, Redis, миграции, Fiber, роуты.

Слои:

```
router.go (composition root модуля: repo -> service -> handler, роуты)
  -> handler (HTTP, типизированные роуты через pkg/httpx)
    -> service (бизнес-правила, объявляет интерфейсы)
      -> repository (pgx, только SQL)
```

Зависимости идут только вниз. Сервис не импортирует Fiber, репозиторий не
знает про HTTP. Каждый модуль — вертикальный срез
(`internal/modules/<name>/`) со своим `router.go`; `internal/app/routes.go`
только вызывает `Register` на публичном или protected-роутере.

## Пакеты

| Пакет | Ответственность |
| --- | --- |
| `pkg/httpx` | типизированные роуты + реестр для OpenAPI |
| `pkg/openapi` | рантайм-OpenAPI 3.0.3 + Scalar UI (ассеты внутри бинаря) |
| `pkg/apperror` | реестр ошибок, конструкторы, единая граница логов/ответов |
| `pkg/logger` | настройка slog, request-scoped логгер, seam под репортер |
| `pkg/jsonx` | `encoding/json/v2`: marshal/unmarshal, строгий парсинг DTO |
| `pkg/jwtx` | JWT sign/parse, извлечение токена (cookie, затем заголовок) |
| `pkg/postgres` | pgxpool, healthcheck, `WithTx` |
| `pkg/redisx` | Redis-клиент, healthcheck |
| `pkg/validx` | общий валидатор, маппинг в `VALIDATION_FAILED` |
| `pkg/response` | хелперы успешных ответов |

## Путь запроса

1. Middleware `RequestID` выдаёт/пробрасывает `X-Request-ID`.
2. `RequestLogger` создаёт request-scoped логгер и кладёт его в контекст.
3. `recover`, `cors`, `helmet` (и `pprof`, если включён).
4. Роуты монтируются под `API_BASE_PATH` (по умолчанию `/api/v1`). `/auth/*` и
   доки — публичные; всё, что требует авторизации, монтируется в protected-группу
   (`middleware.Auth`): сначала cookie `access_token`, затем
   `Authorization: Bearer`. В Fiber групповая middleware защищает только роуты,
   зарегистрированные после неё, поэтому публичные роуты и доки идут первыми.
5. Типизированный handler: строгий парсинг JSON -> валидация -> сервис.
6. Ошибки возвращаются наверх и один раз логируются/форматируются в
   `apperror.Handler`.

## Ошибки

Два вида: `API` (4xx, безопасное сообщение) и `Runtime` (5xx, generic-ответ,
cause — только в логах). Коды — константы в `pkg/apperror/codes.go`, набор
минимальный и стабильный; HTTP-статусы берутся только из `fiber.Status*`.
Успешные ответы идут через `pkg/response` и оборачивают payload в
`{"result": ...}`; ошибки приходят в конверте `error`, поэтому клиент различает
их, не глядя на статус-код.

Поля лога ошибки: `type`, `message`, `trace`, `kind`, `http_status`,
`request_id`, `user_id`, для runtime — `error` (cause).

## Конфигурация

`.env` для локальной разработки (реальные env-переменные приоритетнее),
`.env.example` всегда синхронен. `internal/config.Load()` декодирует и
валидирует всё, включая прод-инварианты (сильный JWT-секрет,
`COOKIE_SECURE=true`, CORS без `*`). Списки через запятую парсятся в
типизированные слайсы.

## Миграции

`migrations/*.sql` (goose Up/Down) встраиваются в бинарь и применяются на
старте при `DB_AUTO_MIGRATE=true` под session advisory lock PostgreSQL — можно
запускать несколько инстансов. Ручные `task migrate:*` остаются для откатов и
разработки.

## Контракт API

Роуты регистрируются только через `pkg/httpx`; реестр питает рантайм-OpenAPI
(`{API_BASE_PATH}/openapi.json`, `{API_BASE_PATH}/docs`). `docs/openapi.json` защищён golden-тестом, таблица
роутов — snapshot-тестом, в CI работает `oasdiff breaking`. Случайное изменение
публичного контракта валит пайплайн.

## Документация

`docs/<topic>/{en,ru}.md`; ADR — `docs/adr/NNNN-<slug>/{en,ru}.md`.
`README.md` — только английский и краткий.
