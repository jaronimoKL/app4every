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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	// Делаем email необязательным — пользователи могут регистрироваться только с логином.
	// PostgreSQL разрешает несколько NULL в UNIQUE-колонке.
	_, err = pool.Exec(ctx, `ALTER TABLE users ALTER COLUMN email DROP NOT NULL;`)
	if err != nil {
		// Игнорируем ошибку если колонка уже nullable
		_ = err
	}

	return pool, nil
}
