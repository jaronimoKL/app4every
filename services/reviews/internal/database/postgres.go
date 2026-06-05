package database

import (
	"context"
	"fmt"
	"time"

	"app4every/services/reviews/internal/config"
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
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Таблица рецензий
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS reviews (
			id           BIGSERIAL PRIMARY KEY,
			user_id      BIGINT NOT NULL,
			title        VARCHAR(500) NOT NULL,
			content_type VARCHAR(20) NOT NULL DEFAULT 'movie',
			status       VARCHAR(20) NOT NULL DEFAULT 'planned',
			rating       SMALLINT CHECK (rating IS NULL OR (rating >= 1 AND rating <= 10)),
			notes        TEXT NOT NULL DEFAULT '',
			poster_url   TEXT NOT NULL DEFAULT '',
			created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate reviews table: %w", err)
	}

	// Таблица ссылок — CASCADE удаляет ссылки вместе с рецензией
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS review_links (
			id         BIGSERIAL PRIMARY KEY,
			review_id  BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
			label      VARCHAR(200) NOT NULL DEFAULT '',
			url        TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate review_links table: %w", err)
	}

	// Индексы
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_review_links_review_id ON review_links(review_id);`)

	return pool, nil
}
