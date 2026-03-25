# Wishlist App

Приложение для управления списками желаний на основе **Clean Architecture**.  
Создано на базе Go backend-шаблона, содержит REST API для работы с вишлистами и их элементами.

![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)
![License](https://img.shields.io/badge/license-MIT-green)

## 🚀 Быстрый старт

```bash
# 1. Запустите PostgreSQL
docker-compose up -d postgres

# 2. Примените миграции
make migrate-up

# 3. Запустите сервер
make run

# 4. Откройте frontend
# Откройте web/index.html в браузере
```

**Готово!** Приложение доступно на `http://localhost:8081`

📖 Подробная инструкция: [START.md](START.md)

## ✨ Возможности

### Backend (Go)
-  REST API для вишлистов и элементов
-  Clean Architecture (Entity → DTO → UseCase → Repository → Controller)
-  PostgreSQL с миграциями
-  Валидация данных
-  CORS поддержка
-  Structured logging
-  Graceful shutdown
-  Docker support

### Frontend (Vanilla JS)
-  Создание и управление вишлистами
-  Добавление элементов с ценой, ссылкой, приоритетом
-  Отметка купленных элементов
-  Публичные/приватные списки
- Современный UI с анимациями
-  Адаптивный дизайн

### Функции вишлистов
-  Название и описание
-  Публичные/приватные списки
-  Привязка к пользователю
-  Автоматические timestamps

### Функции элементов
-  Название и описание
-  URL ссылка на товар
-  Цена
-  Приоритет (0-10)
-  Статус покупки
-  Сортировка по приоритету

## 📚 Документация

- [START.md](START.md) - Быстрый запуск за 4 шага
- [QUICKSTART.md](QUICKSTART.md) - Подробная инструкция по установке
- [USAGE.md](USAGE.md) - Руководство по использованию
- [API_EXAMPLES.md](API_EXAMPLES.md) - Примеры API запросов
- [FEATURES.md](FEATURES.md) - Полное описание возможностей
- [CHANGELOG.md](CHANGELOG.md) - История изменений
- [INSTALL_MAKE.md](INSTALL_MAKE.md) - Установка Make на Windows

## Содержание

- [Технологии](#технологии)
- [Архитектура](#архитектура)
- [Структура проекта](#структура-проекта)
- [Быстрый старт](#быстрый-старт)
- [Конфигурация](#конфигурация)
- [Команды Makefile](#команды-makefile)
- [Миграции](#миграции)
- [API](#api)
- [OAuth2](#oauth2)
- [Swagger](#swagger)
- [Docker](#docker)
- [Как добавить новый домен](#как-добавить-новый-домен)

---

## Технологии

| Назначение | Библиотека |
|---|---|
| HTTP-роутер | [`go-chi/chi`](https://github.com/go-chi/chi) |
| HTTP-рендеринг ответов | [`go-chi/render`](https://github.com/go-chi/render) |
| База данных | [`jackc/pgx`](https://github.com/jackc/pgx) |
| Query-builder | [`Masterminds/squirrel`](https://github.com/Masterminds/squirrel) |
| Сканирование строк БД | [`georgysavva/scany`](https://github.com/georgysavva/scany) |
| Миграции | [`pressly/goose`](https://github.com/pressly/goose) |
| Конфигурация | [`joho/godotenv`](https://github.com/joho/godotenv) |
| Валидация | [`go-playground/validator`](https://github.com/go-playground/validator) |
| Form-декодинг | [`go-playground/form`](https://github.com/go-playground/form) |
| CLI | [`spf13/cobra`](https://github.com/spf13/cobra) |
| UUID | [`google/uuid`](https://github.com/google/uuid) |
| Утилиты для срезов | [`samber/lo`](https://github.com/samber/lo) |
| OAuth2 | [`golang.org/x/oauth2`](https://pkg.go.dev/golang.org/x/oauth2) |
| Swagger | [`swaggo/swag`](https://github.com/swaggo/swag) + [`swaggo/http-swagger`](https://github.com/swaggo/http-swagger) |
| Тесты | [`stretchr/testify`](https://github.com/stretchr/testify) |

---

## Архитектура

Проект следует принципам **Clean Architecture**: каждый слой зависит только от слоя ниже, а не выше. Бизнес-логика изолирована от инфраструктуры.

```
┌─────────────────────────────────┐
│         controller/http         │  ← HTTP-хендлеры, middleware
│   (знает только об usecase)     │
├─────────────────────────────────┤
│            usecase              │  ← Бизнес-логика
│   (знает только об entity/dto)  │
├─────────────────────────────────┤
│           repository            │  ← Работа с БД
│   (реализует интерфейс usecase) │
├─────────────────────────────────┤
│         entity / dto            │  ← Доменные типы, не знают ни о чём
└─────────────────────────────────┘
```

**Главное правило**: зависимости направлены только вниз. `controller` знает об `usecase`, `usecase` не знает о `controller`. `repository` не знает об `usecase`.

**Общение между слоями только через интерфейсы.** Например, `usecase` не знает что репозиторий использует PostgreSQL он знает только об интерфейсе `Repository`. Это позволяет легко подменять реализацию (например, для тестов).

---

## Структура проекта

```
.
├── cmd/
│   ├── app/main.go          ← точка входа HTTP-сервера
│   └── cli/main.go          ← точка входа CLI-утилиты
├── internal/
│   ├── app/
│   │   └── app.go           ← сборка зависимостей (DI), запуск сервера
│   ├── config/
│   │   └── config.go        ← загрузка конфигурации из .env / Vault
│   ├── database/
│   │   └── database.go      ← pgxpool + squirrel builder
│   ├── logger/
│   │   └── logger.go        ← slog с JSON-форматом
│   ├── entity/
│   │   └── article.go       ← доменные сущности (чистые Go-структуры)
│   ├── dto/
│   │   └── article.go       ← input/filter структуры для usecase
│   ├── definitions/
│   │   ├── errors.go        ← sentinel-ошибки (ErrNotFound, ErrForbidden...)
│   │   └── constants.go     ← константы пагинации и прочее
│   ├── usecase/
│   │   ├── contracts.go     ← публичный интерфейс ArticleUseCase
│   │   └── article/
│   │       ├── interfaces.go ← интерфейс Repository (что нужно от БД)
│   │       └── usecase.go   ← реализация бизнес-логики
│   ├── repository/
│   │   └── article/
│   │       ├── repo.go      ← реализация запросов к PostgreSQL
│   │       ├── queries.go   ← squirrel query-builders
│   │       └── mapper.go    ← преобразование DB-строк в entity
│   ├── controller/
│   │   └── http/
│   │       ├── middleware/  ← recoverer, json content-type, auth
│   │       └── v1/
│   │           ├── router.go      ← регистрация маршрутов
│   │           ├── article.go     ← хендлеры статей
│   │           ├── auth.go        ← хендлеры OAuth2
│   │           ├── request/       ← структуры входящих запросов
│   │           └── response/      ← структуры ответов + маппинг ошибок
│   ├── oauth/
│   │   ├── provider.go      ← интерфейс Provider, фабрика New()
│   │   ├── github.go        ← реализация GitHub OAuth2
│   │   └── state.go         ← CSRF state store (in-memory)
│   ├── cli/
│   │   ├── root.go          ← корневая cobra-команда
│   │   └── commands/        ← healthz, migrate:up/down/status/create
│   ├── helpers/
│   │   ├── context.go       ← типизированный доступ к context
│   │   ├── validator.go     ← синглтон validator + form decoder
│   │   ├── slice.go         ← обёртки над samber/lo
│   │   └── ptr.go           ← ToPtr / FromPtr дженерики
│   └── testhelpers/
│       └── helpers.go       ← утилиты для тестов
├── migrations/
│   └── 00001_create_articles.sql  ← goose-миграции
├── docs/swagger/            ← автогенерируемая документация (make swag)
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .env.example
```

---

## Быстрый старт

### Требования

- Go 1.24+
- Docker + Docker Compose
- [golangci-lint](https://golangci-lint.run/usage/install/)
- [swag](https://github.com/swaggo/swag#getting-started)

### Локально

```bash
# 1. Клонировать репозиторий
git clone https://github.com/KaoriEl/golang-boilerplate.git
cd golang-boilerplate

# 2. Скопировать конфиг и заполнить переменные
cp .env.example .env

# 3. Поднять PostgreSQL
docker-compose up -d postgres

# 4. Применить миграции
make migrate-up

# 5. Запустить сервер
make run

# 6. Открыть frontend в браузере
# Откройте файл web/index.html в браузере
# или запустите простой HTTP-сервер:
# python -m http.server 8080 --directory web
# и откройте http://localhost:8080
```

Сервер будет доступен на `http://localhost:8081`.  
Swagger UI: `http://localhost:8081/swagger/index.html`  
Frontend: откройте `web/index.html` в браузере

### Через Docker Compose

```bash
# Собрать и запустить всё
make docker-up

# Остановить
make docker-down

# Пересобрать после изменений
make docker-rebuild
```

---

## Конфигурация

Все параметры задаются через переменные окружения. При локальной разработке они читаются из файла `.env`.

> В production Vault Agent записывает секреты в файлы `/vault/secrets/CRED_<NAME>`.  
> Код сначала пытается прочитать оттуда, при неудаче из переменной окружения.  
> Не пугайся, если видишь в логах предупреждения о том, что секреты не найдены. Это нормально, если ты не используешь Vault.

Скопируй `.env.example` в `.env` и заполни:

```dotenv
# Приложение
APP_ENV=local          # окружение: local | development | production
APP_PORT=8081          # порт HTTP-сервера
LOG_LEVEL=debug        # уровень логов: debug | info | warn | error

# База данных (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=boilerplate
DB_USERNAME=postgres
DB_PASSWORD=postgres

# OAuth2 (GitHub)
# github тут указан для примера, в проекте лучше использовать например VK
OAUTH_PROVIDER=github
OAUTH_CLIENT_ID=your_client_id
OAUTH_CLIENT_SECRET=your_client_secret
OAUTH_REDIRECT_URL=http://localhost:8080/api/v1/auth/callback
```

---

## Команды Makefile

```bash
make run              # запуск сервера локально (go run)
make build            # сборка бинаря → bin/app
make test             # тесты с race-детектором
make lint             # запуск golangci-lint
make lint-fix         # автоисправление линтером
make swag             # регенерация Swagger-документации

make migrate-up       # применить все pending миграции
make migrate-down     # откатить последнюю миграцию
make migrate-status   # показать статус всех миграций
make migrate-create   # создать новый файл миграции (запросит имя)

make docker-build     # сборка Docker-образа
make docker-up        # запуск через docker-compose
make docker-down      # остановка docker-compose
make docker-rebuild   # пересборка и перезапуск

make help             # список всех команд
```

---

## Миграции

Управление миграциями через [goose](https://github.com/pressly/goose) v3.

### Создать новую миграцию

```bash
make migrate-create
# > Migration name: add_users_table
#  Migration created: add_users_table
```

Будет создан файл `migrations/00002_add_users_table.sql`.

### Формат файла миграции

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email      VARCHAR(255) NOT NULL UNIQUE,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);
COMMENT ON TABLE users IS 'Пользователи системы';
COMMENT ON COLUMN users.email IS 'Email адрес пользователя';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

### Правила

- Именование файлов: `00001_snake_case_description.sql` (5 цифр + описание)
- Один файл одна логическая миграция, содержит и `Up` и `Down`
- Всегда использовать `IF NOT EXISTS` / `IF EXISTS` для идемпотентности
- Добавлять `COMMENT ON TABLE` и `COMMENT ON COLUMN` на русском
- Создавать индексы для полей, по которым будет фильтрация или сортировка
- `Down` должен полностью отменять всё что делает `Up`

---

## API

Base URL: `http://localhost:8080/api/v1`

### Вишлисты

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/wishlists` | Список вишлистов (фильтры: `user_id`, `is_public`, пагинация: `page`, `per_page`) |
| `POST` | `/wishlists` | Создать вишлист |
| `GET` | `/wishlists/{id}` | Получить вишлист по UUID |
| `PATCH` | `/wishlists/{id}` | Обновить вишлист (partial update) |
| `DELETE` | `/wishlists/{id}` | Удалить вишлист |

### Элементы вишлиста

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/wishlists/{wishlist_id}/items` | Список элементов вишлиста (фильтр: `is_purchased`, пагинация) |
| `POST` | `/wishlists/{wishlist_id}/items` | Добавить элемент в вишлист |
| `GET` | `/wishlists/{wishlist_id}/items/{id}` | Получить элемент по UUID |
| `PATCH` | `/wishlists/{wishlist_id}/items/{id}` | Обновить элемент (partial update) |
| `DELETE` | `/wishlists/{wishlist_id}/items/{id}` | Удалить элемент |

### Postman коллекция

Для тестирования API можно импортировать коллекцию в Postman:

`postman/Wishlist App.postman_collection.json`


### Формат ошибок

Все ошибки возвращаются в едином формате:

```json
{
  "error": "not found",
  "code": 404
}
```

---

## Линтер

Для поддержания качества кода используется `golangci-lint` с набором принятых у нас линтеров, обязательно запускай `make lint` перед коммитом.

---

## OAuth2

Шаблон содержит готовую структуру для OAuth2-авторизации через GitHub, она указана как пример, но ты можешь легко добавить любой другой провайдер (VK, Google, Facebook и т.д.)


### Добавить нового провайдера

1. Создай файл `internal/oauth/<name>.go`
2. Реализуй интерфейс `Provider`:
   ```go
   type Provider interface {
       Config() *oauth2.Config
       GetUserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error)
   }
   ```
3. Добавь `case` в `internal/oauth/provider.go`:
   ```go
   case "myProvider":
       return newMyProvider(cfg), nil
   ```
4. Установи `OAUTH_PROVIDER=myProvider` в `.env`

---

## Swagger

Swagger UI доступен по адресу: `http://localhost:8081/swagger/index.html`

### Регенерация документации

После изменения аннотаций в хендлерах:

```bash
make swag
```

### Аннотации

Пример аннотации хендлера:

```go
// Create godoc
// @Summary     Создать статью
// @Tags        articles
// @Accept      json
// @Produce     json
// @Param       body body request.CreateArticleRequest true "Данные статьи"
// @Success     201 {object} response.ArticleResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /articles [post]
func (h *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) { ... }
```

---

## Docker

### Dockerfile

Образ использует `golang:1.24-alpine` и запускает приложение через `go run`.  
Подходит для разработки и учебных целей.

```bash
# Собрать образ
make docker-build

# Запустить контейнер с базой данных
make docker-up
```

### docker-compose.yml

Поднимает два сервиса: `app` и `postgres`.  
Данные PostgreSQL сохраняются в volume `postgres_data`.

---

## Как добавить новый домен
> Домен это самостоятельная бизнес-сущность со своей логикой, репозиторием и хендлерами. Например: User, Comment, Order. Каждый домен живёт в своих пакетах по всем слоям архитектуры.

Пошаговый пример добавления домена `User` (пользователи).

### 1. Entity

```go
// internal/entity/user.go
type User struct {
    ID        uuid.UUID
    Email     string
    Name      string
    CreatedAt time.Time
}
```

### 2. DTO

```go
// internal/dto/user.go
type CreateUserInput struct {
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name"  validate:"required,min=1"`
}
```

### 3. Usecase интерфейс репозитория

```go
// internal/usecase/user/interfaces.go
type Repository interface {
    Create(ctx context.Context, input dto.CreateUserInput) (entity.User, error)
    GetByEmail(ctx context.Context, email string) (entity.User, error)
}
```

### 4. Usecase бизнес-логика

```go
// internal/usecase/user/usecase.go
type UseCase struct {
    repo Repository
    log  *slog.Logger
}

func New(repo Repository, log *slog.Logger) *UseCase { ... }
func (uc *UseCase) Create(ctx context.Context, input dto.CreateUserInput) (entity.User, error) { ... }
```

### 5. Публичный контракт

```go
// internal/usecase/contracts.go добавить интерфейс
type UserUseCase interface {
    Create(ctx context.Context, input dto.CreateUserInput) (entity.User, error)
    GetByEmail(ctx context.Context, email string) (entity.User, error)
}
```

### 6. Repository

```go
// internal/repository/user/mapper.go  DB-строки ↔ entity
// internal/repository/user/queries.go squirrel builders
// internal/repository/user/repo.go    реализация Repository
```

### 7. Controller

```go
// internal/controller/http/v1/user.go хендлеры
// internal/controller/http/v1/request/user.go
// internal/controller/http/v1/response/user.go
```

### 8. Зарегистрировать маршруты в router.go

```go
userHandler := newUserHandler(userUC, log)
r.Route("/users", func(r chi.Router) {
    r.Get("/", userHandler.List)
    r.Post("/", userHandler.Create)
})
```

### 9. Миграция

```bash
make migrate-create
# > Migration name: create_users
```

### 10. Собрать зависимости в app.go

```go
userRepo := userrepo.New(db.Pool)
userUC   := useruc.New(userRepo, log)
router   := v1.NewRouter(wishlistUC, userUC, oauthProvider, log)
```

---

## Логирование

Логгер основан на стандартном `log/slog` с JSON-форматом.  
Каждая запись содержит поля `timestamp`, `severity`, `service`, `stage`, `source`.
Поле `rest` обязательное. В него передаётся суть записи: что именно произошло. Не
оставляй его пустым и не дублируй в него технические поля, они уже есть в структуре лога.

```json
{
  "timestamp": "2026-03-14T12:00:00Z",
  "severity": "INFO",
  "source": {"function": "app.Run", "file": "app.go", "line": 42},
  "rest": "starting HTTP server",
  "stage": "local",
  "service": "golang-boilerplate",
  "addr": ":8080"
}
```

---

## Тесты

```bash
# Запустить все тесты
make test

# Конкретный пакет
go test ./internal/dto/...
```

В `internal/testhelpers` лежат утилиты для тестов:

```go
ctx := testhelpers.NewTestContext(t)
testhelpers.AssertNoError(t, err)
testhelpers.AssertError(t, err, definitions.ErrNotFound)
```