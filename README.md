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
wishlist-backend/
│
├── cmd/                          # Точки входа
│   ├── app/main.go              # Основное приложение
│   └── cli/main.go              # CLI команды
│
├── internal/                     # Внутренняя логика
│   ├── app/                     # Инициализация приложения
│   ├── config/                  # Конфигурация
│   ├── controller/http/         # HTTP контроллеры
│   │   ├── middleware/          # Промежуточное ПО
│   │   └── v1/                  # API v1
│   │       ├── auth.go          # Аутентификация
│   │       ├── wishlist.go      # Вишлисты
│   │       └── friendship.go    # Друзья
│   │
│   ├── usecase/                 # Бизнес-логика
│   │   ├── auth/                # Логика авторизации
│   │   ├── wishlist/            # Логика вишлистов
│   │   └── friendship/          # Логика дружбы
│   │
│   ├── repository/              # Работа с БД
│   │   ├── user/                # Пользователи
│   │   ├── wishlist/            # Вишлисты
│   │   └── friendship/          # Дружеские связи
│   │
│   └── entity/                  # Модели данных
│       ├── user.go
│       └── wishlist.go
│
├── migrations/                   # Миграции БД
└── docker-compose.yml           # Оркестрация контейнеров

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

Посмотреть все роли в системе:

docker exec wishlist_postgres psql -U postgres -d wishlist -c "SELECT id, name, description FROM roles;"

Посмотреть все пользователи и их роли:

docker exec wishlist_postgres psql -U postgres -d wishlist -c "
SELECT u.username, u.email, r.name as role 
FROM users u 
LEFT JOIN user_roles ur ON u.id = ur.user_id AND ur.is_active = true 
LEFT JOIN roles r ON ur.role_id = r.id 
ORDER BY u.username;"

