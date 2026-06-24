package database

import (
	"context"
	"fmt"
	"time"

	"app4every/services/auth/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(cfg *config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSslMode)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Автомиграция первой таблицы пользователей
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate users table: %w", err)
	}

	// Добавляем колонку username, если её ещё нет (безопасно для повторного запуска).
	// PostgreSQL разрешает несколько NULL в UNIQUE-колонке, поэтому
	// существующие строки без username не нарушат ограничение.
	_, err = pool.Exec(ctx, `
		ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100) UNIQUE;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to add username column: %w", err)
	}

	// Очищаем пустые/null email и делаем email обязательным (NOT NULL)
	_, _ = pool.Exec(ctx, `UPDATE users SET email = 'user_' || id || '@example.com' WHERE email IS NULL OR email = '';`)
	_, err = pool.Exec(ctx, `ALTER TABLE users ALTER COLUMN email SET NOT NULL;`)
	if err != nil {
		_ = err
	}

	// Таблица связей дружбы
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS friendships (
			id         BIGSERIAL PRIMARY KEY,
			user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			friend_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			status     VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, friend_id)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate friendships table: %w", err)
	}

	// Индексы для friendships
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_friendships_user_id ON friendships(user_id);`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_friendships_friend_id ON friendships(friend_id);`)

	return pool, nil
}
