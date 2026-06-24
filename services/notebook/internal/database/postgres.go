package database

import (
	"context"
	"fmt"
	"time"

	"app4every/services/notebook/internal/config"
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

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Автомиграция таблицы заметок
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS notes (
			id         BIGSERIAL PRIMARY KEY,
			user_id    BIGINT NOT NULL,
			title      VARCHAR(500) NOT NULL DEFAULT '',
			content    TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate notes table: %w", err)
	}

	// Индекс для быстрой выборки заметок пользователя
	_, err = pool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create notes index: %w", err)
	}

	return pool, nil
}
