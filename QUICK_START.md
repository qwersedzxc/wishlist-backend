# ⚡ Быстрый старт Wishlist App

## Первый запуск на новом ПК

### 1. Установить Docker Desktop
https://www.docker.com/products/docker-desktop

### 2. Клонировать проект
```bash
git clone https://github.com/YOUR_USERNAME/wishlist-app.git
cd wishlist-app
```

### 3. Настроить Backend

```bash
cd backend

# Скопировать пример конфигурации
copy .env.example .env

# Отредактировать .env - указать свои ключи:
# - OAUTH_CLIENT_ID и OAUTH_CLIENT_SECRET (GitHub)
# - S3_ACCESS_KEY_ID и S3_SECRET_ACCESS_KEY (Yandex Cloud)
# - SMTP_USERNAME и SMTP_PASSWORD (Email)
```

### 4. Настроить Frontend

```bash
cd frontend

# Скопировать пример конфигурации
copy .env.example .env

# .env уже настроен правильно, ничего менять не нужно
```

### 5. Запустить всё

```bash
# Backend (PostgreSQL + API + Nginx)
cd backend
docker-compose up -d

# Frontend (React app)
cd ../frontend
docker-compose up -d --build
```

### 6. Открыть приложение
```
http://localhost
```

## Что нужно настроить в GitHub OAuth

1. Зайти в https://github.com/settings/developers
2. Создать новое OAuth приложение или отредактировать существующее
3. Указать:
   - **Homepage URL:** `http://localhost`
   - **Authorization callback URL:** `http://localhost/api/v1/auth/oauth/callback`
4. Скопировать Client ID и Client Secret в `backend/.env`

## Остановка сервисов

```bash
# Backend
cd backend
docker-compose down

# Frontend
cd frontend
docker-compose down
```

## Просмотр логов

```bash
# Backend
cd backend
docker-compose logs -f backend

# Frontend
cd frontend
docker-compose logs -f frontend

# PostgreSQL
cd backend
docker-compose logs -f postgres
```

## Troubleshooting

### База данных не создается
```bash
docker exec -it wishlist_postgres psql -U postgres -c "CREATE DATABASE wishlist;"
```

### Порт 80 занят
В `backend/docker-compose.yml` измените:
```yaml
nginx:
  ports:
    - "8080:80"  # Вместо 80:80
```

### Пересоздать всё с нуля
```bash
cd backend
docker-compose down -v
docker-compose up -d

cd ../frontend
docker-compose down
docker-compose up -d --build
```
