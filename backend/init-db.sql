-- Создание базы данных wishlist (если не существует)
SELECT 'CREATE DATABASE wishlist'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'wishlist')\gexec

-- Подключение к базе wishlist
\c wishlist;

-- Создание расширений если нужно
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";