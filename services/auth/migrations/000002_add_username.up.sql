-- Миграция: добавление колонки username в таблицу users
-- Файл для документации. Фактически выполняется в database/postgres.go при старте.
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100) UNIQUE;
