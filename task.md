# Backend Template — план реализации

Цель: production-ready шаблон на Go, одинаково работающий на Windows и macOS, с чистой слоистой структурой, JWT-авторизацией, PostgreSQL/Redis и полной AI-обвязкой (rules + skills).

Стек: Go 1.27+ · Fiber v3.5 · pgx/v5 · Redis (go-redis/v9) · PostgreSQL · JWT HS256 · `encoding/json/v2`.

---

## 1. Структура проекта

```text
template-golang/
├── cmd/
│   └── api/
│       └── main.go                  # единственная точка входа (-check-env флаг)
├── internal/                        # код приложения (не импортируется извне)
│   ├── app/
│   │   ├── app.go                   # сборка зависимостей (DI), lifecycle, graceful shutdown
│   │   └── routes.go                # единая регистрация роутов и middleware
│   ├── config/
│   │   ├── config.go                # типизированный Load() + валидация (fail-fast)
│   │   └── config_test.go
│   ├── modules/                     # доменные вертикальные срезы: один роут = один модуль
│   │   └── auth/
│   │       ├── router.go            # composition root модуля: repo -> service -> handler + роуты
│   │       ├── handler.go           # HTTP: парсинг/валидация запроса, вызов сервиса
│   │       ├── service.go           # бизнес-логика, интерфейсы зависимостей
│   │       ├── repository.go        # SQL через pgx, реализация интерфейсов
│   │       ├── model.go             # доменные структуры с db:"..." тегами
│   │       └── dto.go               # request/response структуры
│   ├── middleware/
│   │   ├── auth.go                  # JWT: cookie -> Authorization, skip /auth/*
│   │   └── ctx.go                   # user_id и request-scoped logger в контексте
│   └── (realtime/ — только при добавлении realtime по скиллу add-realtime; по умолчанию отсутствует)
├── pkg/                             # переиспользуемые пакеты без доменной логики
│   ├── postgres/                    # pgxpool init, healthcheck, WithTx helper
│   ├── redisx/                      # redis client init, healthcheck
│   ├── jwtx/                        # Claims{UserID, SignAt}, Sign/Parse, extractor
│   ├── apperror/
│   │   ├── codes.go                 # реестр констант: kind, HTTP-статус, public message
│   │   ├── error.go                 # Error, API()/Runtime()/Wrap(), trace, лимит message 120
│   │   └── report.go                # единая точка логирования ошибок на границе
│   ├── logger/                      # slog-обёртка: JSON/text, FromCtx, request_id, Reporter-seam
│   ├── jsonx/                       # encoding/json/v2: строгий Unmarshal, streaming, адаптеры Fiber
│   ├── httpx/                       # типизированные роуты Get/Post[Req,Resp] + реестр роутов
│   ├── openapi/                     # OpenAPI 3.0.3 из реестра + reflection DTO (без аннотаций)
│   ├── response/                    # единый JSON-ответ/ошибка
│   └── validx/                      # singleton validator, Fiber-биндинг
├── migrations/                      # SQL-миграции + автозапуск
│   ├── embed.go                     # //go:embed *.sql + goose.Up на старте
│   └── <version>_<name>.sql         # goose Up/Down
├── tools/
│   ├── tools.go                     # pinned dev-tools (goose CLI) через blank import
│   └── genmodule/                   # генератор module: handler+service+repo+dto
├── docs/                            # документация: docs/<topic>/{en,ru}.md
│   ├── openapi.json                 # golden-спека (генерируется из кода, язык-агностична)
│   ├── architecture/
│   │   ├── en.md
│   │   └── ru.md
│   ├── skills/
│   │   ├── en.md
│   │   └── ru.md
│   ├── error-codes/                 # контракт кодов ошибок для фронта
│   │   ├── en.md
│   │   └── ru.md
│   └── adr/0001-<slug>/
│       ├── en.md
│       └── ru.md
├── .opencode/                       # AI-слой проекта: правила и скиллы в одном месте
│   ├── instructions/                # правила качества кода
│   │   ├── go.instructions.md
│   │   ├── project.instructions.md
│   │   ├── tests.instructions.md
│   │   └── performance.instructions.md
│   └── skills/
│       ├── add-module/SKILL.md
│       ├── pgx-queries/SKILL.md
│       ├── create-migration/SKILL.md
│       ├── error-handling/SKILL.md
│       ├── tdd-tests/SKILL.md
│       ├── add-realtime/SKILL.md
│       ├── openapi/SKILL.md
│       ├── commit/SKILL.md
│       ├── preflight/SKILL.md
│       └── add-skill/SKILL.md
├── .github/
│   └── workflows/ci.yml             # только CI; AI-правил здесь нет
├── AGENTS.md                        # короткий операционный хаб AI: маршрутизация задач на скиллы
├── README.md                        # всегда EN, краткий: требования, запуск, ссылка на docs
├── LICENSE                          # MIT по умолчанию; при форке — правообладатель/год
├── opencode.json                    # конфиг opencode (skills, instructions, permissions)
├── .env.example                     # полный список переменных с dev-дефолтами
├── .golangci.yml
├── .air.toml                        # hot reload
├── Taskfile.yml                     # кросс-платформенный раннер (вместо Makefile)
├── .commitlint.yaml                 # Conventional Commits (Go-линтер, без Node)
├── Dockerfile                       # multi-stage: golang -> distroless nonroot
├── .dockerignore
├── docker-compose.yml               # postgres + redis (+ app в профиле full)
├── .editorconfig
├── .gitattributes                   # LF, чтобы Windows не ломал форматирование
├── .vscode/                         # общий debug-профиль (коммитим)
│   ├── launch.json                  # F5: API под dlv, envFile=.env
│   ├── settings.json                # gopls/gofumpt, единый стиль для команды
│   └── extensions.json              # рекомендация golang.go
├── go.mod
└── task.md
```

Правила структуры:

- Добавить фичу = создать `internal/modules/<name>/` + одну строчку вызова `Register` в `internal/app/routes.go`.
- **`router.go` — composition root модуля**: сам создаёт `repository -> service -> handler`, принимает только нужную инфраструктуру через `Deps` (pool, redis, config, signer, logger) и регистрирует свои роуты через `pkg/httpx`. `internal/app/routes.go` не знает внутренностей модуля — только вызывает `Register(public|protected, registry, deps)`.
- Публичные модули получают `public := root.Group("")`, защищённые — `protected := root.Group("/", middleware.Auth(signer))`; модуль регистрирует полные пути, чтобы реестр OpenAPI совпадал с реальными роутами.
- `handler` не содержит SQL и бизнес-логики; `service` не знает про Fiber; `repository` не знает про HTTP.
- Интерфейсы объявляются там, где используются (в `service.go`), реализации — в `repository.go`.
- `internal/` — код приложения, `pkg/` — только переиспользуемое и без доменных понятий.
- Никаких глобальных singleton и `init()`: всё собирается в `app.go` и прокидывается явно.
- Модуль = один домен, 4–6 файлов. Всё общее выносится в `pkg/`, а не копируется.
- **Локальных утилит нет**: запрещены `utils.go`, `helpers.go`, `common.go`, `misc.go` в модулях и пакетах. Любой переиспользуемый примитив — это отдельный пакет в `pkg/` с именем по зоне ответственности (`jwtx`, `jsonx`, `validx`), а не свалка функций.
- Пакет `pkg/*` = одна ответственность; если в пакете появляются две несвязанные темы — он делится.

---

## 2. Правила и скиллы (AI-ready)

### 2.1 Instructions (правила качества)

Все AI-артефакты живут только в `.opencode/` — `.github` предназначен исключительно для CI и не используется как место хранения правил.

| Файл | Содержимое |
| --- | --- |
| `.opencode/instructions/go.instructions.md` | Копия [awesome-copilot/go.instructions.md](https://github.com/github/awesome-copilot/blob/main/instructions/go.instructions.md) (с указанием источника; не редактируем) |
| `.opencode/instructions/project.instructions.md` | Архитектура слоёв, зоны ответственности, DRY/KISS/SOLID, наследование и полиморфизм, правила регистрации модулей, работа с БД, ошибки, конфиг, запреты |
| `.opencode/instructions/tests.instructions.md` | TDD (red-green-refactor), только stdlib `testing`, table-driven, рукописные fakes, контрактные тесты `pkg/*`, coverage gate, benchmarks, build-тег `integration` |
| `.opencode/instructions/performance.instructions.md` | Фокус на аллокации: preallocate, reuse, streaming, `sync.Pool`; бенчмарки и профилирование hot paths |

Ключевые пункты `project.instructions.md`:

1. **БД**: только pgx/v5; каждый маппинг строк — через `db:"col"` теги и `pgx.CollectRows(rows, pgx.RowToStructByName[T])` (или `RowToStructByNameLax` для опциональных колонок). Ручной `rows.Scan` и ORM запрещены. Только параметризованные запросы, `context.Context` первым аргументом.
2. **Слои**: `router.go` (composition root модуля) -> handler -> service -> repository, зависимости только вниз, возврат ошибок с `%w`; `internal/app/routes.go` только вызывает `Register` модулей.
3. **Ошибки**: только через `pkg/apperror`; код (`type`) — исключительно константа из реестра `codes.go` (никаких строковых литералов); HTTP-статус — только константы `fiber.Status*`/`http.Status*` (никаких чисел вроде `404`); вид `API` (4xx) или `Runtime` (5xx); message ≤ 120 символов; для 5xx наружу — generic-сообщение, cause — только в лог. Логирование ошибки — один раз на границе (error handler), не в каждом слое.
4. **Минимум констант ошибок**: набор кодов маленький и стабильный; перед добавлением нового кода — переиспользовать существующий (детализацию несут `message` и данные, а не новый `type`); новый код создаётся только если фронт обязан по-разному на него реагировать, и это невозможно выразить существующими. Лишние константы не делают ошибку понятнее и усложняют валидацию на фронте; удаление/переименование кода — breaking change для API.
5. **Контракт API заморожен**: роуты, пути, методы, `operationId`, DTO-поля, коды ошибок и HTTP-статусы — публичный контракт. Меняются только по явному запросу на конкретную ручку/DTO; «заодно» — запрещено. Задача «измени модуль X» не даёт права трогать чужие модули, `routes.go`, `pkg/*` или форматы ответов. Перед завершением задачи — `task api:check` (diff OpenAPI + snapshot роутов), при осознанном breaking-изменении golden-файлы обновляются явным коммитом с пометкой `BREAKING CHANGE`.
6. **Логи**: только `pkg/logger` и `logger.FromCtx(ctx)`; структурированные поля, секреты/PII не попадают в лог.
7. **JSON**: только `encoding/json/v2` (Go 1.27) через `pkg/jsonx`; входящие DTO — строго с `RejectUnknownMembers(true)`; `encoding/json` v1 и сторонние JSON-библиотеки — запрещены.
8. **TDD**: сначала тест на контракт (падающий), потом реализация; багфикс начинается с теста, воспроизводящего баг; `pkg/*` покрыт unit-тестами как стабильный API.
9. **Аллокации**: hot paths пишутся с фокусом на минимум аллокаций (preallocate, переиспользование, streaming); оптимизация — только после профиля и бенчмарка.
10. **HTTP**: все ответы через `pkg/response`; валидация входа через `validx`; новый роут — типизированный хелпер из `pkg/httpx` (method/path/DTO фиксируются в реестре) + вызов в `routes.go`; схемы для OpenAPI генерируются автоматически, комментарии в хендлерах не пишутся.
11. **Конфиг**: только через `internal/config`; новый env — обязательно в `.env.example` + типизированное поле + валидация.
12. **Запрещено**: new dependency без обоснования, `any` без необходимости, глобальное состояние, cgo, локальные `utils/helpers/common`, логирование и возврат ошибки одновременно.
13. **Документация**: `README.md` — всегда на английском, краткий (требования, быстрый старт, ссылка на `docs/`); остальная документация — паттерн `docs/<topic>/en.md` + `docs/<topic>/ru.md`; изменение делается сразу в обеих версиях, README — единственное исключение.
14. **AI-скиллы**: до написания кода классифицировать задачу по таблице маршрутизации в `AGENTS.md` и загрузить релевантный скилл; после правок скиллов синхронизировать `docs/skills/{en,ru}.md` и таблицу маршрутизации.
15. **Коммиты**: Conventional Commits (`commitlint` на Go, без Node), один смысловой шаг — один коммит; коммит создаётся только когда пользователь попросил или агент явно обозначил точку коммита (скилл `commit`).

#### Принципы организации кода (входит в `project.instructions.md`)

**DRY — Don't Repeat Yourself**

- Один факт/правило — в одном месте: схема — в миграциях, SQL — в repository, коды ошибок — в реестре `codes.go`, env — в `internal/config`.
- Дублируется знание, а не строки: если правило меняется в двух местах не синхронно — это баг.
- Правило трёх: выносим в `pkg/` после третьего повторения или как только код нужен второму модулю. Спекулятивный DRY (абстракция «на будущее») запрещён — это преждевременная абстракция.
- Проверка на ревью: «если поменяется X, сколько файлов правим?». Больше одного — знание не централизовано.

**KISS — Keep It Simple, Stupid**

- Простейшее работающее решение: stdlib раньше зависимости, явный код раньше магии, линейный код раньше хитрой абстракции.
- Ранние `return`, happy path слева, минимум вложенности (совпадает с go.instructions).
- Ничего «на будущее»: интерфейс появляется при второй реализации или необходимости мока, а не заранее.
- Конфиг без слоёв, DI руками в `app.go`, роутинг без метапрограммирования.
- Исключение — hot path: решение может быть неочевидным, но обязано быть локальным, изолированным в `pkg/*` и покрыто бенчмарком.

**SOLID (как это работает здесь)**

- **S — Single Responsibility**: `handler` — только HTTP, `service` — только бизнес-правила, `repository` — только SQL; `pkg/*` — одна тема; файл — одна причина изменения. Вопрос-проверка: «сколько причин заставит этот тип меняться?» — больше одной: дели.
- **O — Open/Closed**: расширение через новую реализацию/модуль, а не правку чужих `switch`. Примеры: новый код ошибки — добавление в реестр, точки вызова не трогаются; realtime подключается отдельным пакетом по скиллу `add-realtime`, доменные модули не меняются; новый endpoint — новый модуль + строка в `routes.go`.
- **L — Liskov Substitution**: интерфейс в `service` — контракт. Любая реализация (pgx, mock, redis) ведёт себя одинаково: те же ошибки, тот же порядок, никаких паник и скрытых эффектов. Контрактные тесты интерфейса прогоняются на каждой реализации.
- **I — Interface Segregation**: интерфейсы маленькие (1–3 метода) и описывают потребность потребителя (`TokenParser`, `Broadcaster`). God-интерфейсы и «интерфейс на весь модуль» запрещены. Интерфейс объявляется у потребителя, а не у реализации.
- **D — Dependency Inversion**: `service` зависит от интерфейса, а не от pgx/Redis; связывание абстракции с реализацией происходит в composition root модуля (`router.go`), а `internal/app` передаёт модулю только инфраструктуру (`Deps`); поэтому модули тестируются без инфраструктуры.

**Наследование и полиморфизм (в терминах Go)**

- Классического наследования в Go нет; его роль выполняют встраивание (embedding) и композиция: `type Service struct { repo Repository; log *slog.Logger }`. Встраивание — для переиспользования поведения, а не для «иерархий типов».
- Полиморфизм — только через интерфейсы: один вызов сервиса, разные реализации (Postgres/Redis/mock/`noop`), выбор — в `router.go` модуля или `app.go` по конфигу/флагу.
- «Accept interfaces, return concrete types»: интерфейс — только там, где нужна абстракция; наружу — конкретные типы.
- Type switch по `kind` допустим только внутри `pkg/apperror` (маппинг в HTTP/лог); в бизнес-коде ветвлений по конкретным типам нет.
- Композиция вместо «базовых классов»: маленькие типы + встраивание + интерфейсы; `BaseService`-паттерн запрещён.

Эти принципы — часть `.opencode/instructions/project.instructions.md`; скиллы на них ссылаются, а ревью каждого изменения включает проверку по ним.

### 2.2 Skills (`.opencode/skills/<name>/SKILL.md`)

| Skill | Триггер | Что делает |
| --- | --- | --- |
| `add-module` | «добавь endpoint/фичу/модуль» | Пошаговый сценарий по TDD: сначала тесты, затем вертикальный срез `router/handler/service/repository/dto`, парсинг DTO через `jsonx` + `validx`, `router.go` собирает граф и регистрирует роут; в `routes.go` одна строка `Register` |
| `pgx-queries` | «запрос к БД», «новый repository-метод» | Обязательный паттерн `db` тегов + `RowToStructByName`, транзакции через `WithTx`, минимум аллокаций в выборках, правила миграций |
| `create-migration` | «миграция», «изменить схему» | `goose` naming, up/down, индексы, автозапуск на старте, запрет destructive-изменений без обратной миграции |
| `error-handling` | «новая ошибка», «лог», «обработка ошибок» | Реестр констант, HTTP-статусы только из `fiber.Status*`, политика минимализма (переиспользовать, не плодить), kind API/Runtime, лимит message 120, trace-флаг, логирование на границе, без секретов в логах |
| `tdd-tests` | «тесты», «TDD», «покрытие», «бенчмарк» | Red-green-refactor только на stdlib `testing`, table-driven, рукописные fakes, контрактные тесты `pkg/*`, coverage gate, benchmarks (`-benchmem`) |
| `add-realtime` | «добавь realtime», «нужен websocket», «ws-слой» | Создаёт WS-слой по конвенциям: пакет `internal/realtime`, зависимость contrib websocket, флаг, роут, auth через `jwtx`, тесты, обновление docs — см. 4.4 |
| `openapi` | «swagger», «openapi», «orval», «схема api» | Типизированный роут-хелпер, генерация спеки из reflection, `/docs` (Scalar), golden `docs/openapi.json`, oasdiff — см. 4.11 |
| `commit` | «закоммить», «commit», «сделай коммит» | Conventional Commits: scope = модуль, императив, ≤72 символа, `BREAKING CHANGE` при ломающем контракте; прогон `task commit:lint` (Go-commitlint); используется только по явному запросу или в явной точке коммита |
| `docs` | «документация», «обнови доки», «docs», «readme», «adr» | Создание и правка `docs/<topic>/{en,ru}.md` и ADR: структура, синхронность языков, README только EN, обновление `docs/skills/{en,ru}.md` и `docs/error-codes/{en,ru}.md` |
| `preflight` | «готово», «проверь», «перед продом», «сделай ревью» | Definition of Done: `task verify`, `task api:check` (контракт не изменился), тесты/покрытие, доки в обеих версиях, корневые файлы (`.gitignore`/README/LICENSE), `.env.example`, реестр ошибок, миграции, бенчмарки — без этого задача не считается выполненной |
| `add-skill` | «добавь скилл», «обнови правила для AI» | Как написать SKILL.md (what+when в description), зарегистрировать в `opencode.json`, таблица маршрутизации в `AGENTS.md`, обновить `docs/skills/{en,ru}.md` |

#### Слои AI-обвязки: что загружается всегда, а что по требованию

| Слой | Где | Когда загружается | Содержимое | Чего там быть НЕ должно |
| --- | --- | --- | --- | --- |
| Навигация | `AGENTS.md` | всегда, короткий (≤1 экран) | стек, команды, карта структуры, таблица «задача -> скилл(ы)», обязательный цикл работы, ссылки | детальные правила, шаблоны, процедуры |
| Правила | `.opencode/instructions/*.md` | всегда (через `opencode.json`) | жёсткие константы проекта: слои, БД, ошибки, JSON, тесты, аллокации, запреты | пошаговые процедуры и код-шаблоны |
| Процедуры | `.opencode/skills/*/SKILL.md` | по требованию (модель сама вызывает `skill`) | полный сценарий: шаги, команды, настройки, шаблоны, примеры, локальный DoD | общие правила, дублирующие instructions |

- `AGENTS.md` — операционный хаб: его задача — понять «что и как делать дальше и какой скилл дернуть», а не пересказывать проект целиком. Всё исполнение — в скиллах.
- `docs/skills/{en,ru}.md` — человекочитаемое описание скиллов (что, когда, как обновлять), синхронно с таблицей в `AGENTS.md`.
- `opencode.json`: `$schema`, `instructions` (явный список из 4 файлов), `skills.paths: [".opencode/skills"]`, `permission` для безопасных дефолтов.
- Ни одно AI-правило не должно лежать в `.github/` — это слой CI/CD, а не слой правил.

#### Автовыбор и качество скиллов

- `description` каждого `SKILL.md` — формула «что делает + когда срабатывает» с триггерными словами (рус+англ); модель видит описания всегда и вызывает `skill` сама, без ручного указания.
- Скилл самодостаточен: шаги, команды, настройки, шаблоны и локальный чек-лист — после загрузки не нужно ничего додумывать.
- Цепочки выполнения: `add-module` -> `tdd-tests`, `pgx-queries`; `create-migration` -> `pgx-queries`; изменения API -> `openapi`; перед «готово» -> `preflight`.
- `commit` вызывается только явно (пользователь просит или агент обозначил точку коммита); `docs` — при любом создании/правке документации.
- `add-skill` держит систему целостной: регистрация в `opencode.json`, таблица маршрутизации, `docs/skills/{en,ru}.md`, проверка, что триггеры не пересекаются.
- Проверка в DoD: дать агенту 3 типовые задачи (endpoint, миграция, ошибка) — до кода он должен сам загрузить соответствующий скилл; `preflight` не даёт сказать «готово» без проверок.

### 2.3 Линт и форматирование

- `golangci-lint` (v2-конфиг): `govet`, `staticcheck`, `errcheck`, `revive`, `gocritic`, `bodyclose`, `contextcheck`, `errorlint`, `exhaustive`, `prealloc`, `perfsprint`, `makezero`, `misspell`, `noctx`, `tparallel`; форматтеры `gofmt` + `gofumpt` + `goimports`.
- `go vet`, `go test ./...` — обязательны; всё это в `task check` и в CI.
- `.editorconfig` + `.gitattributes (eol=lf)` — единый стиль на Windows/macOS.

---

## 3. Зависимости и инструменты

### Runtime-пакеты

| Пакет | Версия | Зачем |
| --- | --- | --- |
| `github.com/gofiber/fiber/v3` | v3.5+ | HTTP-сервер |
| `github.com/jackc/pgx/v5` (+`pgxpool`) | v5 | PostgreSQL, маппинг по `db`-тегам |
| `github.com/redis/go-redis/v9` | v9 | Redis-клиент |
| `github.com/golang-jwt/jwt/v5` | v5 | JWT sign/parse |
| `github.com/joho/godotenv` | latest | загрузка `.env` (не перетирает реальные env) |
| `github.com/sethvargo/go-envconfig` | latest | типизированный парсинг env + `required`/`default` |
| `github.com/go-playground/validator/v10` | v10 | валидация DTO |
| `github.com/pressly/goose/v3` | v3 | миграции (SQL + embed) |

Стандартная библиотека (без зависимостей):

- `encoding/json/v2` + `encoding/json/jsontext` (Go 1.27) — JSON: строгий парсинг и streaming; обёртки в `pkg/jsonx`, см. 4.7.
- `uuid` (Go 1.27) — UUID вместо `github.com/google/uuid` (из ТЗ про random UID и `user_id`).
- `log/slog`, `errors` (включая `errors.AsType`), `crypto/*` — базис.

Тесты: только stdlib `testing` — unit, fuzz, golden и benchmarks (`testing.B.Loop`, `-benchmem`). Сторонние assert/mock-библиотеки не тянем: сравнения через `reflect`/`slices`/`maps`, моки — руками (маленькие fakes, реализующие интерфейс потребителя). Интеграционные тесты — тег `integration`, те же `testing` + DSN из env (`TEST_POSTGRES_DSN`, `TEST_REDIS_ADDR`), поднятые docker-compose/CI-сервисами.

Встроенные middleware Fiber v3 использовать по максимуму (не писать своё): `requestid`, `recover`, `logger`, `cors`, `helmet`, `limiter`, `healthcheck`, `timeout`, `pprof` (под флагом). Собственные middleware — только `auth` и прокидывание user_id в контекст.

### Dev-инструменты (не в runtime)

- `go-task` (`Taskfile.yml`) — кросс-платформенный запуск задач: `setup`, `dev` (air), `build`, `test`, `test:cover`, `bench`, `lint`, `check`, `verify`, `api:check`, `openapi:dump`, `commit:lint`, `migrate:up/down/create`, `gen:module`, `env:check`, `docker:build`, `docker:run`.
- `air-verse/air` — hot reload (Windows/macOS).
- `goose` — через `tools/tools.go` (pinned), запускается `go run`.
- `oasdiff` (Go-бинарь) — через `tools/tools.go`, проверка breaking-изменений OpenAPI в CI.
- `commitlint` — Go-реализация (`github.com/conventionalcommit/commitlint`, pinned в `tools/tools.go`), конфиг `.commitlint.yaml`; commit-msg hook ставится `commitlint init` (кросс-платформенно, без Node), CI проверяет диапазон коммитов.
- `dlv` (Delve) — отладочный адаптер для VSCode: ставится `task setup` через `go install github.com/go-delve/delve/cmd/dlv@<pin>` (вне go.mod, чтобы не тянуть в модуль), работает на Windows/macOS.
- `golangci-lint`, Docker + `docker-compose` (postgres, redis).

### Сознательно НЕ берём

- ORM (GORM и пр.) — есть `db`-теги pgx.
- Viper/koanf — конфиг только `.env` + `internal/config`.
- zap/logrus — `log/slog` из stdlib (см. 4.5).
- sonic/jsoniter/easyjson/goccy — выбран stdlib `encoding/json/v2` (Go 1.27), см. 4.7.
- testify/testcontainers/mockery — только stdlib `testing`; интеграционные тесты через реальные сервисы из docker-compose/CI.
- `github.com/google/uuid` — в Go 1.27 есть stdlib `uuid`.
- Sentry/OTel как зависимость — только интерфейс `logger.Reporter` (noop), чтобы не тащить в шаблон.
- Makefile как основной раннер — не работает на Windows по умолчанию.

---

## 4. Прочее (не потерять при реализации)

### 4.1 Авторизация — фиксируем ТЗ

- JWT (HS256), claims: `user_id` (uuid), `sign_at` (unix), плюс `iat`/`exp`. Секрет и TTL — из env (`JWT_SECRET`, `JWT_TTL`).
- Извлечение токена строго по порядку: cookie `access_token` -> заголовок `Authorization: Bearer <token>`. Нет ни того, ни другого / токен невалиден / истёк -> `401` в формате `pkg/apperror`, без редиректов.
- Middleware не применяется к `/auth/*` (публичная группа роутов; в `routes.go` два группы: public `/auth` и protected с middleware).
- `user_id` кладётся в контекст запроса; helper в `internal/middleware/ctx.go`.
- `POST /auth/mini-apps/telegram` — заглушка: генерирует случайный `uid`, выпускает токен, ставит HttpOnly-cookie (`Secure` в prod, `SameSite=Lax`) и возвращает `{"token": "...", "user_id": "..."}`. Место для будущей реальной валидации Telegram `initData` — изолировано в `service.go`.
- Cookie-параметры (`COOKIE_NAME`, `COOKIE_DOMAIN`, `COOKIE_SECURE`) — из env.

### 4.2 .env: список переменных и правила

```dotenv
APP_ENV=development          # development|staging|production
APP_NAME=template-golang

LOG_LEVEL=info               # debug|info|warn|error
LOG_FORMAT=auto              # auto|json|text (auto: text в dev, json в prod)
ERROR_TRACE_ENABLED=true     # 0/1, default 1 — добавлять trace (file.go:line) в ошибки
ERROR_STACKTRACE_ENABLED=false  # 0/1 — полный stack для 5xx/PANIC (включать на staging)

HTTP_HOST=0.0.0.0
HTTP_PORT=3001
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s
HTTP_IDLE_TIMEOUT=60s
SHUTDOWN_TIMEOUT=10s

POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/app?sslmode=disable
POSTGRES_MAX_CONNS=10
POSTGRES_MIN_CONNS=1
DB_AUTO_MIGRATE=true         # миграции применяются при старте приложения (goose + advisory lock)

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=dev-secret-change-me   # required, в prod обязателен не-дефолтный
JWT_TTL=24h
COOKIE_NAME=access_token
COOKIE_DOMAIN=
COOKIE_SECURE=false

CORS_ALLOWED_ORIGINS=http://localhost:3000,https://app.example.com  # несколько доменов через запятую

FEATURE_OPENAPI_ENABLED=true  # /docs + /openapi.json (Scalar UI); в prod — false или под auth
FEATURE_PPROF_ENABLED=false
```

Правила:
- Всё, что меняется между окружениями, и все feature-flags — только в `.env` (флаги с префиксом `FEATURE_`, всегда bool, дефолт `false`).
- `.env` в `.gitignore`, `.env.example` — всегда актуален и содержит безопасные dev-дефолты.
- **Списки в env — через запятую**: `CORS_ALLOWED_ORIGINS=a,b,c` парсится в `internal/config` в `[]string` (trim пробелов, пропуск пустых, валидация формата origin). В коде списки уже типизированы — `strings.Split` в бизнес-логике запрещён.
- Проверка перед стартом: `config.Load()` валидирует обязательные поля, типы и инварианты (prod: секрет не дефолтный, `COOKIE_SECURE=true`, CORS без `*`) и падает с понятным списком ошибок. `task env:check` и флаг `api -check-env` — тот же код для CI и pre-start.
- Реальные env-переменные имеют приоритет над `.env` (godotenv без override).
- Новую переменную нельзя добавить, не обновив `.env.example` + `internal/config` (это правило для AI-скиллов).

### 4.3 Кросс-платформенность и Docker

- Только pure-Go зависимости, `CGO_ENABLED=0` — простой build на обеих ОС.
- `Taskfile.yml` вместо Makefile; задачи не зависят от bash/PowerShell-специфики.
- Только `filepath` для путей, LF через `.gitattributes`, UTF-8 через `.editorconfig`.
- Docker Compose для инфраструктуры; Air для hot reload; CI-матрица `ubuntu/macos/windows` (build + unit-тесты).
- Интеграционные тесты с Docker — только на ubuntu/локально, тег `integration`.

**Dockerfile (multi-stage, маленький и безопасный образ):**

- Builder: `golang:1.27-alpine`, кэш зависимостей отдельным слоем (`go mod download` до копирования кода), `CGO_ENABLED=0`, `-trimpath`, `-ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT"`; ARG `TARGETOS/TARGETARCH` под multi-arch (`docker buildx` amd64+arm64).
- Runtime: `gcr.io/distroless/static-debian12:nonroot` (CA-сертификаты и tzdata уже внутри, нет shell — меньше поверхность атаки), `USER nonroot`, `EXPOSE`, `ENTRYPOINT ["/app"]`.
- `HEALTHCHECK CMD ["/app", "-healthcheck"]` — встроенный флаг той же точки входа дергает локальный `/healthz` и возвращает код выхода (не тянем curl/wget в образ).
- `.dockerignore`: `.git`, `.env*` (кроме примера), `bin/`, `tmp/`, `coverage*`, `docs/`, `.opencode/`, `*.md`, `docker-compose*`, `Taskfile.yml` — контекст минимальный.
- `docker-compose.yml` получает сервис `app` (profile `full`) с `depends_on: postgres/redis (service_healthy)`; дефолтный профиль — только инфраструктура.
- Валидация после реализации: `task docker:build` (сборка), `docker run --env-file .env -p 3001:3001` + `curl /healthz` и `/readyz`; в CI — build образа и smoke-запуск контейнера на ubuntu.
- `.env` в образ не копируется: конфиг приходит только через env/секреты окружения (12-factor).

### 4.4 Realtime (WebSocket) — по умолчанию отсутствует, добавляется по запросу

Базовый шаблон не содержит WebSocket-кода и зависимости `gofiber/contrib/websocket`. Но правила и скилл `add-realtime` постоянно лежат в проекте, чтобы при запросе «добавь realtime/websocket» всё создалось по конвенциям, без пересмотра архитектуры.

Инварианты, которые соблюдаются всегда (даже когда кода ещё нет) — фиксируются в `project.instructions.md`:

- Realtime — отдельный транспортный пакет `internal/realtime`, изолированный от домена: бизнес-модули импортируют только интерфейс, библиотеку websocket — никто, кроме `internal/realtime`.
- Интерфейс публикации (`Broadcaster`) объявляет потребитель (service модуля); `router.go` модуля (или `app.go` для общей инфраструктуры) связывает его с реализацией, как и любую другую зависимость.
- События — типизированные структуры в dto модуля, а не `map[string]any`; сериализация через `jsonx`.
- Авторизация WS — тем же `jwtx`: cookie, а для браузерного клиента `?token=` (WebSocket не умеет передавать заголовки); неавторизованный апгрейд — `401`/`UNAUTHORIZED`.
- Публикация событий не блокирует HTTP-запрос: буферизованный канал + `select` с `default`, никаких горутин без владельца.
- Аллокации: переиспользуемые буферы/кадры в `sync.Pool`, broadcast не дублирует payload на каждого клиента.
- `Hub.Close()` обязательно вызывается в graceful shutdown; течет goroutine — баг.

Что делает скилл `add-realtime` (когда реально понадобится):

1. Добавляет зависимость `github.com/gofiber/contrib/v3/websocket` (v3-модуль) в `go.mod`.
2. Создаёт `internal/realtime/` (`hub.go` + тесты), реализующий `Broadcaster`/`Subscriber`; при `FEATURE_WS_ENABLED=false` роут не регистрируется.
3. Добавляет `FEATURE_WS_ENABLED=false` в `.env.example` и поле в `internal/config` с валидацией (флаг появляется вместе с кодом, а не заранее).
4. Регистрирует апгрейд-роут в `routes.go` условно + auth-through-`jwtx`.
5. Покрывает тестами на stdlib `testing` (hub с фейковыми коннектами, ping/pong, авторизация) и, при необходимости, JSON-стриминг через `jsontext`.
6. Обновляет `docs/architecture/{en,ru}.md`, таблицу маршрутизации в `AGENTS.md` и, при расширении, `docs/skills/{en,ru}.md`.

Триггер скилла: «добавь realtime», «нужен websocket», «сделай ws-слой» — агент сам находит его по описанию/таблице маршрутизации.

### 4.5 Логирование

- `log/slog` из stdlib + обёртка `pkg/logger`. Формат: `LOG_FORMAT=auto` (text в development, JSON в prod/staging), уровень — `LOG_LEVEL`.
- Базовые поля записи: `time`, `level`, `msg`, `service`, `env`, `version` (из build-инфо через `-ldflags`); для запросов добавляются `request_id`, `method`, `path`, `status`, `latency_ms`, `user_id`.
- Request-scoped logger создаётся middleware и кладётся в `context.Context`; логировать во всех слоях только через `logger.FromCtx(ctx)`. Глобального логгера нет.
- Один поток — stdout, одна строка на запись (12-factor, Docker/Loki/ELK); ротация — задача внешней системы, не приложения.
- Access-логи — Fiber `logger` middleware: 5xx -> error, 4xx -> warn, остальное -> info.
- Секреты и PII (Authorization, cookie, DSN, пароли, тела запросов) в лог не попадают — только whitelist полей.
- Интерфейс `logger.Reporter` (по умолчанию `noop`) — задел под внешний репортер (Sentry/OTel) без зависимости сейчас.

### 4.6 Обработка ошибок

Два вида ошибок:

| kind | Зона | HTTP | Уровень лога | Что уходит клиенту |
| --- | --- | --- | --- | --- |
| `API` | клиент | 4xx | warn | безопасный message |
| `Runtime` | сервер/инфра | 5xx | error | generic message, без деталей |

Формат записи об ошибке (ровно эти три поля + служебные):

- `type` — константа ошибки. Константы живут только в `pkg/apperror/codes.go`, строковые литералы кодов запрещены. Реестр `map[Code]Info{Kind, HTTPStatus, PublicMessage}` — единственный источник правды.
- `HTTPStatus` — только константы Fiber/`net/http` (`fiber.StatusBadRequest`, `fiber.StatusUnauthorized`, `fiber.StatusNotFound`, `fiber.StatusConflict`, `fiber.StatusUnprocessableEntity`, `fiber.StatusTooManyRequests`, `fiber.StatusRequestEntityTooLarge`, `fiber.StatusInternalServerError` и т.п.). Числовые литералы статусов запрещены настолько, что проверяются тестом (`codes.go` не содержит цифр статусов).
- Группы кодов (минимальный стартовый набор, дальше только по политике ниже):
  - API: `VALIDATION_FAILED`, `UNAUTHORIZED`, `TOKEN_EXPIRED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `RATE_LIMITED`, `PAYLOAD_TOO_LARGE`.
  - Runtime: `INTERNAL`, `DB_QUERY_FAILED`, `DB_CONNECT_FAILED`, `DB_MIGRATION_FAILED`, `REDIS_FAILED`, `HTTP_CLIENT_FAILED`, `PANIC`, `CONFIG_INVALID`.
- **Политика минимализма (важно для фронта)**: один код покрывает все однотипные ситуации — конкретику несут `message` и `details`, а не новый `type`. Новый код добавляется только если клиент обязан обрабатывать ситуацию отдельной веткой (например, `TOKEN_EXPIRED` -> обновить токен) и это нельзя выразить существующими. Переименование/удаление кода — breaking change. Реестр — публичный контракт, отражён в `docs/error-codes/{en,ru}.md`.
- `message` — max 120 символов. Лимит обеспечивается базовым конструктором (его вызывают `API`/`Runtime`/`Wrap`): обрезка + тест на длину, а не дисциплина разработчика.
- `trace` — путь до места ошибки (`pkg/.../file.go:123`, для runtime дополнительно имя функции). Захватывается через `runtime.Caller` в конструкторе при `ERROR_TRACE_ENABLED=1`; при `0` поле не пишется (экономия CPU на hot path).
- Служебные: `kind`, `http_status`, `request_id`, `user_id`; для Runtime — `error` (cause), только в серверном логе.

Как используется:

- Конструкторы: `apperror.API(code, msg)`, `apperror.Runtime(code, cause)`, `apperror.Wrap(cause, code)`; поддержка `errors.Is/As`; в модулях не создаются ad-hoc ошибки с новыми кодами — код передаётся в конструктор из реестра.
- Логирование — один раз на границе, в Fiber ErrorHandler (`pkg/apperror/report.go`), не в каждом слое (см. правила 3–4 в 2.1).
- `recover` middleware превращает panic в `PANIC`; при `ERROR_STACKTRACE_ENABLED=1` в лог добавляется полный stack, иначе — только trace.
- Тесты реестра: у каждого кода есть Info, коды уникальны, `HTTPStatus` — валидная константа (без чисел), статус соответствует kind; тест ловит неиспользуемые и дублирующиеся коды; `docs/error-codes/{en,ru}.md` синхронизирован с реестром (golden-тест).

Для прода заложено, но выключено по умолчанию:

- `ERROR_STACKTRACE_ENABLED` — полный stack для 5xx/PANIC (включать на staging; в `.env.example` default 0).
- Сэмплирование/дедупликация повторяющихся ошибок, чтобы не залить лог-хранилище штормом 5xx.
- Счётчик метрик по `code` (задел под Prometheus, под feature-flag).
- `version`/`commit` в каждой записи об ошибке для быстрого сопоставления с релизом.
- Внешний Reporter (Sentry/OTel) — точки вызова уже есть в `report.go`.

### 4.7 JSON-обработка (encoding/json/v2)

- Go 1.27: `encoding/json/v2` — новый стандарт, `encoding/json` (v1) внутри переведён на v2-движок. Новый код пишется только на v2 API (`Marshal`/`Unmarshal` с variadic `Options`).
- `pkg/jsonx` — единственное место работы с JSON: обёртки `Marshal`/`Unmarshal` с дефолтными опциями проекта, строгий парсер входа и адаптеры под Fiber (`JSONEncoder`/`JSONDecoder`) — сигнатуры v2 вариативные, в `fiber.Config` напрямую не присваиваются.
- В `app.go` Fiber создаётся с `JSONEncoder: jsonx.FiberEncoder`, `JSONDecoder: jsonx.FiberDecoder`; все `c.JSON()` и `c.Bind()` идут через v2.
- Входящие DTO парсим строго: `jsonv2.Unmarshal(body, &dto, jsonv2.RejectUnknownMembers(true))` — неизвестное поле даёт `VALIDATION_FAILED` (API, 400). Ловит опечатки в интеграциях и мусор от клиентов.
- Аллокации: для больших тел — `jsonv2.UnmarshalRead`/`MarshalWrite` и `jsontext`-кодеки вместо промежуточных `[]byte`; буферы ответов преаллоцируются.
- Стриминг для WebSocket/SSE — `encoding/json/jsontext` (`Encoder`/`Decoder` на токенах), без буферизации всего сообщения.
- Исходящие структуры — `json:"name"` теги (`omitempty`/`omitzero` для optional); `db` и `json` теги на одной структуре не смешиваем (model — для БД, dto — для API).
- jsonb в pgx: pgx использует stdlib `encoding/json` (v2-движок); если нужны явные v2-опции — кастомный JSONB-codec регистрируется в `pkg/postgres`.
- Порядок обработки запроса: `jsonx` (парсинг) -> `validx` (валидация) -> service; ошибки обоих этапов — `API`/`VALIDATION_FAILED`.

### 4.8 Производительность и аллокации

- Фокус на аллокации — часть definition of done для hot paths: middleware (auth, logger), JSON-обработка, repository, WS-бродкасты.
- Обязательные приёмы:
  - преаллокация: `make([]T, 0, n)`, `slices.Grow`, `make(map[K]V, n)` при известном размере.
  - строки: `strings.Builder` (с `Grow`), `strconv.Append*` вместо `fmt.Sprintf` в hot paths.
  - минимум конверсий `string <-> []byte`; API на `[]byte` там, где это уместно; `unsafe` запрещён.
  - `sync.Pool` для буферов и тяжёлых временных структур (JSON, WS-кадры); обязательный сброс перед возвратом в пул.
  - маленькие структуры — по значению, большие/мутируемые — по указателю.
  - `slog`: атрибуты собираются только при включённом уровне (`Enabled`), статичные `slog.Attr` переиспользуются.
  - Fiber: `c.Body()`/`c.Query()` могут ссылаться на внутренние буферы — не удерживать после возврата из хендлера (`Immutable=false`), копировать только осознанно.
  - pgx: `RowToStructByName` без промежуточных `map[string]any`; prepared statements; переиспользование структур при батчах.
- Измерение: `go test -bench=. -benchmem`, `testing.B.Loop` (Go 1.27), pprof под `FEATURE_PPROF_ENABLED`; оптимизация — только после профиля, изменение без цифр «до/после» не принимается.
- `task bench` в Taskfile; в CI — smoke-бенчмарки без hard gate, сравнение результатов через benchstat.
- Правило для AI: любое изменение hot path обязано сохранять или улучшать аллокации и сопровождаться бенчмарком.

### 4.9 Наблюдаемость и безопасность

- `GET /healthz` (liveness, без зависимостей) и `GET /readyz` (ping PG + Redis с таймаутом).
- Graceful shutdown: `signal.NotifyContext` + `app.ShutdownWithTimeout`.
- Единый JSON-формат ответов, CORS, helmet, rate limiter, ограничение размера тела, таймауты сервера.
- Все SQL — параметризованные; секреты не логируются и не коммитятся.

### 4.10 Миграции и тесты

- Миграции — обычные SQL-файлы `migrations/<version>_<name>.sql` (goose: `-- +goose Up/Down`); порядок задаёт версия в имени.
- **Автозапуск при старте**: `migrations/embed.go` встраивает `*.sql` (`//go:embed`), и в `app.go` после подключения PG и до старта HTTP выполняется `goose.Up` (флаг `DB_AUTO_MIGRATE=true` по умолчанию). Ошибка миграции — фатальный стоп приложения с `DB_MIGRATION_FAILED` (fail-fast), сервер не поднимается на несовместимой схеме.
- Multi-instance безопасность: goose с session advisory lock — параллельные инстансы не гоняются за одну миграцию.
- В prod допустимо `DB_AUTO_MIGRATE=false` и применение отдельным шагом; ручные `task migrate:create/up/down/status` остаются для разработки и отката.
- **TDD (red-green-refactor)**: сначала падающий тест на контракт/поведение, затем минимальная реализация, затем рефакторинг. Для багфиксов — сначала тест, воспроизводящий баг.
- **Контрактные тесты `pkg/*`**: каждая экспортируемая функция/тип/ошибка покрыта unit-тестами, включая граничные случаи и ошибки. Цель — 100% контракта, hard gate в CI — не ниже 90% покрытия `pkg/*` (порог только повышается).
- **Стабильность API**: смена exported API `pkg/*` — breaking change, требует обновления тестов и `docs/architecture/{en,ru}.md`; golden-тесты для форматов (error envelope, поля логов, JWT claims).
- Только stdlib `testing`: table-driven + subtests, `t.Helper()`, `t.Cleanup()`; сравнения через `reflect`/`slices`/`maps`; моки — рукописные fakes; fuzz-тесты для парсеров (`jwtx`, `validx`).
- Benchmarks hot paths — stdlib `testing.B` (`B.Loop`, `-benchmem`), рядом с кодом.
- Интеграционные (build-тег `integration`): реальные PostgreSQL/Redis из docker-compose или CI-сервисов, DSN из `TEST_POSTGRES_DSN`/`TEST_REDIS_ADDR`; проверяют маппинг `db`-тегов, автозапуск миграций, auth flow и контракты repository.

### 4.11 OpenAPI/Swagger UI (runtime, без аннотаций)

Да, это возможно без комментариев в хендлерах — за счёт типизированной регистрации роутов. Нужно один раз зафиксировать конвенцию: хендлеры регистрируются через дженерик-хелперы, а спека строится из реестра + reflection по DTO.

- `pkg/httpx`: `httpx.Get[Req, Resp](r, path, handler)`, `Post`, `Patch`, `Delete` — обёртки над `fiber.Router`. Каждый вызов записывает в реестр: метод, путь, `operationId` (из имени хендлера: `auth.Handler.MiniAppTelegram` -> `auth.miniAppTelegram`, без ручного описания), типы `Req`/`Resp`, признак защищённости (protected-группа), tag = имя модуля. В хендлерах нет ни одного комментария/аннотации.
- `pkg/openapi`: на старте обходит реестр и строит OpenAPI **3.0.3** (не 3.1 — для совместимости с Orval) в `map[string]any`:
  - JSON Schema для DTO из reflection: типы, `required` (по `omitempty`/указателям), `enum` (константы), `format` (`uuid`, `date-time`, `uri`), `default`, `nullable`;
  - ответы: успешная схема + единый error envelope, в который подставляются все коды из реестра `apperror` (реестр — единственный источник, расхождение ловится тестом);
  - `securitySchemes`: `cookieAuth` (cookie `access_token`) и `bearerAuth` — для роутов protected-группы; `/auth/*` публичные;
  - `servers`, `info` (название/версия из build-инфо).
- Раздача — всё в рантайме, никакого билд-степа и кодогенерации:
  - `GET /openapi.json` — актуальная спека (строится один раз при старте и кэшируется);
  - `GET /docs` — красивая Scalar UI (тёмная тема, поиск, «Try it»), статика кладётся через `go:embed` в `pkg/openapi` (без CDN-зависимости); фолбэк — Swagger UI, если нужен классический вид;
  - всё под флагом `FEATURE_OPENAPI_ENABLED=true` в dev / `false` в prod (или закрыто auth) — в `.env.example` включено.
- Orval: фронт тянет `http://localhost:3001/openapi.json` (dev) или golden `docs/openapi.json` (CI/офлайн) и генерирует типизированный клиент. Схема стабильна, ломающие изменения видны на ревью.
- Golden-спека для Orval и CI: `task openapi:dump` пишет `docs/openapi.json`; unit-тест сравнивает рантайм-спеку с файлом (несовпадение = забыли обновить контракт); CI прогоняет `oasdiff breaking` против базовой ветки — случайное ломающее изменение API валит пайплайн.
- Автотесты контракта: snapshot роутов (метод+путь+`operationId`+статусы) — «AI переписал модуль» физически не может незаметно поменять ручку: diff покажет изменение и golden-тест упадёт.
- Ограничения честно: генерируются только типы, которые реально используются в сигнатурах `Req`/`Resp`; экзотика (полиморфизм, кастомные сериализаторы) описывается расширением реестра точечно. Если захочется совсем без своего кода — есть готовый Huma (поддерживает Fiber v3, тоже без аннотаций), но он тянет свои конвенции и пачку зависимостей, поэтому в шаблоне выбран тонкий свой слой.
- Экономия для команды: добавление ручки = типизированный хелпер, спека и UI обновились сами после рестарта, комментарии писать не нужно (максимум — имя хендлера, из которого рождается `operationId`).

### 4.12 Документация: `docs/<topic>/{en,ru}.md`

- Единый паттерн: тема — папка, язык — файл. `docs/architecture/en.md` + `docs/architecture/ru.md`, `docs/error-codes/{en,ru}.md`, `docs/skills/{en,ru}.md`, ADR — `docs/adr/NNNN-<slug>/{en,ru}.md`. `docs/openapi.json` — единственный язык-агностичный артефакт.
- `README.md` — **всегда на английском и краткий**: что за проект, требования (Go 1.27+, Docker, dlv для debug), быстрый старт (`task setup`, `task dev`, `F5` в VSCode, миграции), ссылки на `docs/` за деталями, лицензия. Без простыней — глубина живёт в `docs/`.
- Правило синхронности: любое изменение темы делается сразу в `en.md` и `ru.md`, структура заголовков одинаковая (удобно сравнивать), README — единственное исключение (только EN).
- `docs/error-codes/{en,ru}.md` синхронизированы с реестром `apperror` через golden-тест: добавил/переименовал код — обновил обе версии, иначе CI падает.
- ADR — на каждое заметное решение («почему json/v2», «почему минимум кодов ошибок», «почему realtime по умолчанию отсутствует», «почему свой OpenAPI вместо Huma») — это память проекта между сессиями.
- Создание и правка доков — через скилл `docs` (структура, обе версии, ссылки, обновление таблицы скиллов), вручную никто формат не помнит.

### 4.13 Корневые файлы репозитория (обязательно поправить при реализации)

- `.gitignore`: добавить `bin/`, `tmp/`, `coverage.out`, `.env` + `.env.*` (кроме `.env.example`), `.DS_Store`, `Thumbs.db`, `.idea/`, локальные пользовательские настройки; при этом `go.sum`, `.air.toml`, `docs/openapi.json`, `.commitlint.yaml`, `.vscode/launch.json|settings.json|extensions.json` коммитятся — не игнорировать.
- `README.md`: переписать из заглушки в краткий EN-документ по 4.12 (требования, быстрая установка/запуск, ссылки на `docs/`, лицензия). Никаких простыней — детали в `docs/`.
- `LICENSE`: проверить и обновить (тип лицензии осознанно, MIT по умолчанию; год и правообладатель — реальные или явно плейсхолдер). Файл не удалять и не оставлять чужие имена.
- Проверка в `preflight`: эти три файла входят в DoD — если добавлены новые артефакты (например, `.env.local`, локальные бинарники), они должны быть в `.gitignore`; README/LICENSE — соответствовать текущему проекту.

### 4.14 Debug в VSCode и dlv

- В конце реализации запуск проекта проверяется именно из VSCode: `F5` (Start Debugging) / `Ctrl+F5` (Run Without Debugging) — API должен подняться под отладчиком без правок launch.json руками.
- `.vscode/launch.json` (коммитится, один для всей команды):
  - `API (debug)`: `type: go`, `mode: debug`, `program: ./cmd/api`, `cwd: ${workspaceFolder}`, `envFile: ${workspaceFolder}/.env`, `console: integratedTerminal`, `stopOnEntry: false`, `showLog: true`;
  - `API (check-env)`: тот же `program` + `args: ["-check-env"]` — быстрая проверка конфига в дебаггере;
  - `Test current package`: `mode: test`, `program: ${fileDirname}`.
- `dlv` ставится на `task setup` и обновляется пиннингом версии; VSCode-расширение `golang.go` рекомендовано через `.vscode/extensions.json`.
- `.vscode/settings.json`: `gopls` с `formatting.gofumpt`, форматирование при сохранении, отсутствие автообновления инструментов — единый опыт на Windows/macOS/CI.
- `delve` работает с `CGO_ENABLED=0` и не требует изменений кода; если Docker-образ используется как runtime — отладка всё равно локальная (dlv внутри distroless не тащим).
- Проверка в DoD: `task setup` поставил dlv, `F5` поднимает сервер, брейкпоинт в хендлере останавливает выполнение, горячий reload Air не конфликтует с отладчиком (не запускать оба на одном порту).

### 4.15 Порядок реализации (чек-лист)

Статус на момент реализации: всё ниже помечено `[x]`, кроме пунктов, которые
требуют окружения, недоступного в этой среде (Docker/VS Code/go-task/
golangci-lint не установлены — соответствующие проверки нужно прогнать у себя).

- [x] Корневые файлы: `.gitignore` (`.env*` кроме примера, `bin/`, `tmp/`, `coverage.out`, OS/`.idea`; `.vscode` коммитим), `README.md` (краткий EN), `LICENSE` (MIT, правообладатель `module-foundry`)
- [x] `go mod init`, Taskfile, `.editorconfig`, `.gitattributes`, `.air.toml`, `.golangci.yml`
- [x] `internal/config` + `.env.example` + pre-start check (`-check-env`), включая `LOG_FORMAT`, `ERROR_TRACE_ENABLED`, `ERROR_STACKTRACE_ENABLED`, `DB_AUTO_MIGRATE`, `FEATURE_OPENAPI_ENABLED`, парсинг `CORS_ALLOWED_ORIGINS` в `[]string` (проверено: `-check-env` → configuration OK)
- [x] `pkg/logger` (slog, Format/Level, FromCtx, request_id, Reporter-noop) + request-scoped логгер в middleware
- [x] `pkg/jsonx` (encoding/json/v2: Marshal/Unmarshal, `RejectUnknownMembers`, deterministic-режим, адаптеры Fiber) + wiring в `fiber.Config`
- [x] `pkg/httpx` (типизированные роуты + реестр) + `pkg/openapi` (OpenAPI 3.0.3 из reflection, `/openapi.json`, Scalar `/docs` через go:embed) + golden `docs/openapi.json` + `task api:check` (проверено тестами)
- [x] `pkg/apperror`: реестр `codes.go` (HTTP-статусы только `fiber.Status*`, тест «нет чисел»), конструкторы с лимитом message 120 и trace-флагом, `report.go`, ErrorHandler + recover
- [x] `pkg/`: `response`, `validx`, `postgres` (pgxpool + `WithTx` через интерфейс `Beginner`), `redisx`, `jwtx`
- [x] TDD-контрактные тесты для всех `pkg/*` (только stdlib `testing`) + бенчмарки hot paths + coverage gate: 93.2% при пороге 90% (excluded `pkg/postgres`, `pkg/redisx` — их добирает integration-тест)
- [x] `internal/middleware/auth.go` + public/protected группы роутов
- [x] `internal/modules/auth`: `POST /auth/mini-apps/telegram` (заглушка с random UID, cookie + JSON); вертикальный срез с `router.go` (repo -> service -> handler), `app/routes.go` вызывает только `Register`
- [x] Realtime не создаём: инварианты в `project.instructions.md`, скилл `add-realtime` на месте
- [x] `migrations/embed.go` + автозапуск goose при старте (session advisory lock, fail-fast) + `00001_init.sql`
- [x] `tools/genmodule` + `tools/covergate` (проверено smoke-запуском генератора)
- [x] `.vscode/launch.json` + `settings.json` + `extensions.json`; dlv в `task setup` — [ ] проверить F5-запуск и брейкпоинт в VS Code
- [x] `.commitlint.yaml` + Go-commitlint (через `go run @v0.12.0` в Taskfile, без Node) + CI-проверка сообщений
- [x] AI-обвязка: `AGENTS.md` (операционный хаб + маршрутизация), `opencode.json`, 11 скиллов, 4 instructions
- [ ] `Dockerfile` + `.dockerignore` + `task docker:build`: файлы готовы, [ ] провалидировать сборку и smoke в среде с Docker (в CI job `docker`)
- [ ] `golangci-lint` + `task verify` — конфиг и Taskfile готовы; [ ] прогнать `task verify` при установленных go-task/dl/golangci-lint
- [x] Документация: EN `README.md` + `docs/<topic>/{en,ru}.md` (architecture, skills, error-codes, ADR 0001) + скилл `docs` + golden-тест синхронизации error-codes
- [x] Проверка DoD (выполнимое здесь): `go build ./...`, `go vet ./...`, `gofmt` чистые; `go test ./...` зелёный; `/auth/mini-apps/telegram` покрыт тестами; защищённые роуты 401/200 покрыты тестами middleware; `/openapi.json` и golden совпадают; ошибки содержат `type`/`message`/`trace`; неизвестное JSON-поле → `VALIDATION_FAILED`; миграции применяются на старте; docs синхронны; чек-лист скиллов в `AGENTS.md`. [ ] Осталось у себя: Docker smoke, F5, `task verify`, `task dev` на Windows/macOS

---

## 5. Вайбкодинг: риски и усиления

Цель проекта — быстро создать, проработать, запустить и вывести в прод. Ниже — где AI-разработка обычно ломается и что именно в шаблоне это закрывает.

### 5.1 Где обычно ломается

| # | Риск | Как проявляется | Защита в шаблоне |
| --- | --- | --- | --- |
| 1 | Дрейф паттернов между сессиями | один модуль на pgx-тегах, другой на ручном Scan; разные форматы ответов | instructions + skills + таблица маршрутизации, `genmodule`, ADR |
| 2 | «Готово, потому что скомпилировалось» | тесты не запускались, покрытие не проверено, доки не обновлены | `task verify`, скилл `preflight`, DoD в AGENTS.md |
| 3 | Разрастание кодов ошибок и зависимостей | 40 констант, фронт не понимает; новая либа под каждую мелочь | политика минимума кодов, реестр как контракт, dependency policy + `govulncheck` |
| 4 | Локальные утилиты и копипаста | `helpers.go` в модуле, третья копия одного парсера | запрет локальных утилит, правило трёх, всё в `pkg/*` по зонам |
| 5 | Небезопасный код | SQL-склейка, секреты в логах/гите, `CORS=*`, cookie без Secure | pgx-параметризация, apperror без утечек, gosec/gitleaks, prod-валидация конфига |
| 6 | Миграции ломают прод | auto-migrate на нескольких инстансах, нет down, destructive without backup | advisory lock, `DB_AUTO_MIGRATE=false` в prod, expand/contract, бэкап |
| 7 | Скрытая деградация производительности | N+1, лишние аллокации в hot path, копирование тел | perf-rules, benchmarks, pprof, prepared statements |
| 8 | Случайное изменение API | «поправил модуль» — а переписаны роуты, DTO и коды | замороженный контракт в правилах, типизированный реестр роутов, snapshot-тест, golden `docs/openapi.json`, `task api:check`, oasdiff breaking в CI |
| 9 | «Локально работало» | Docker-образ не собирается в CI, env другой | docker smoke в CI, `-check-env` при старте, конфиг только из env |
| 10 | Потеря контекста и решений | через месяц никто не помнит, почему так сделано | ADR, docs en/ru, короткий AGENTS.md |
| 11 | Слишком большой шаг агента | 30 файлов в одном заходе, конфликты и регрессии | маленькие завершённые шаги (один модуль за раз), скиллы по одной зоне ответственности, CI gates |
| 12 | Утечка секретов в промпты/гит | реальный `.env` в контексте или коммите | `.env` в `.gitignore` и `.dockerignore`, только `.env.example`, secret scanning |

### 5.2 Что добавить (усиления)

- `task verify` — единая команда «всё зелёное»: fmt-check, `go vet`, `golangci-lint`, unit-тесты с `-race`, coverage gate `pkg/*`, сборка. Локально и в CI — один источник правды.
- `task api:check` — защита контракта: snapshot роутов + diff сгенерированной OpenAPI против golden `docs/openapi.json`; в CI — `oasdiff breaking` против базовой ветки. Случайно «переписать API» не получится.
- Скилл `preflight` — перед словом «готово» проходит DoD: verify, `api:check`, обе языковые версии доков, `.env.example`, реестр ошибок, миграции, бенчмарки.
- Скилл `commit` — Conventional Commits по Go-commitlint (scope = модуль, `BREAKING CHANGE` при смене контракта); используется только когда пользователь попросил или агент явно обозначил точку коммита. Hook ставится `commitlint init`, CI валидирует сообщения коммитов; Node не нужен.
- Безопасность в CI: `govulncheck ./...`, `gosec` (или через golangci-lint), `gitleaks` (secret scan), сканирование образа (trivy/docker scout) — опционально, но дешево.
- Контракт с фронтом: `docs/error-codes/{en,ru}.md` + golden-тест синхронизации с реестром; OpenAPI 3.0.3 в рантайме + Scalar UI (см. 4.11), Orval тянет спеку по URL или из golden-файла.
- ADR-папка с шаблоном: каждое заметное решение фиксируется (json/v2, минимум кодов ошибок, realtime off by default, auto-migrate, свой OpenAPI вместо Huma).
- Release-дисциплина: semver-теги, `version/commit` через `-ldflags`, Docker-тег `semver+sha`, changelog по желанию.
- Прод-готовность: prod-инварианты конфига (не дефолтные секреты, `COOKIE_SECURE=true`, CORS-список, `DB_AUTO_MIGRATE=false`), health/ready, rate limit, бэкапы и план отката.
- Наблюдаемость: `request_id` во всех логах, счётчик ошибок по `code` (seam), `Reporter` под Sentry/OTel.
- Опционально: devcontainer/`.tool-versions` (mise) — если нужна 1:1 среда Windows/macOS/CI.
- Правила работы для AI: один модуль за раз; массовые рефакторинги только после тестов; никаких новых зависимостей, кодов ошибок и правок публичного контракта без обоснования и явного запроса.

### 5.3 Быстрый путь в прод

1. `task setup && task dev` — локальный старт с docker-compose, миграции применились автоматически.
2. `task verify` — зелёный локально (то же самое прогонит CI).
3. CI: unit (matrix ubuntu/macos/windows), lint, integration, docker smoke, security scans.
4. Прод-конфиг: `APP_ENV=production`, реальные секреты, `COOKIE_SECURE=true`, CORS-список, `DB_AUTO_MIGRATE=false`.
5. Миграции: бэкап -> expand (совместимо со старой версией) -> деплой -> contract (удаление старого — следующим релизом).
6. Образ: buildx multi-arch, тег semver+sha, скан уязвимостей.
7. Смоук после деплоя: `/healthz`, `/readyz`, `/auth/mini-apps/telegram`, формат ошибки.
8. Откат: предыдущий тег образа; down-миграции — только если подготовлены, иначе forward-fix.
