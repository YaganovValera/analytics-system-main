-- 1. Создаём БД analytics, если не существует
SELECT 'CREATE DATABASE analytics'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'analytics'
) \gexec

-- 2. Создаём БД auth, если не существует
SELECT 'CREATE DATABASE auth'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'auth'
) \gexec

-- 3. Включаем расширение timescaledb в БД analytics
\connect analytics

CREATE EXTENSION IF NOT EXISTS timescaledb;
