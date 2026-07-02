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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	// Миграции для новых полей Anime
	_, _ = pool.Exec(ctx, `ALTER TABLE reviews ADD COLUMN IF NOT EXISTS shikimori_id INTEGER;`)
	_, _ = pool.Exec(ctx, `ALTER TABLE reviews ADD COLUMN IF NOT EXISTS tmdb_id INTEGER;`)
	_, _ = pool.Exec(ctx, `ALTER TABLE reviews ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';`)
	_, _ = pool.Exec(ctx, `ALTER TABLE reviews ADD COLUMN IF NOT EXISTS episodes_total INTEGER;`)
	_, _ = pool.Exec(ctx, `ALTER TABLE reviews ADD COLUMN IF NOT EXISTS current_episode INTEGER NOT NULL DEFAULT 0;`)
	_, _ = pool.Exec(ctx, `ALTER TABLE reviews ADD COLUMN IF NOT EXISTS aniliberty_alias TEXT NOT NULL DEFAULT '';`)
	_, _ = pool.Exec(ctx, `ALTER TABLE reviews ADD COLUMN IF NOT EXISTS shikimori_score NUMERIC(4,2);`)

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

	// Таблица жанров — CASCADE удаляет жанры вместе с рецензией
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS review_genres (
			id        BIGSERIAL PRIMARY KEY,
			review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
			name      VARCHAR(100) NOT NULL
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate review_genres table: %w", err)
	}

	// Индексы
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_review_links_review_id ON review_links(review_id);`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_review_genres_review_id ON review_genres(review_id);`)

	// ── Миграции для групповых списков ──
	
	// Безопасное создание таблицы users (если auth-service еще не запустил миграцию)
	// чтобы внешние ключи REFERENCES users(id) не падали при первом старте.
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id            BIGSERIAL PRIMARY KEY,
			email         VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			username      VARCHAR(100) UNIQUE,
			created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to safe-migrate users table: %w", err)
	}

	// 1. Группа (список)
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS group_lists (
			id          BIGSERIAL PRIMARY KEY,
			name        VARCHAR(200) NOT NULL,
			owner_id    BIGINT NOT NULL REFERENCES users(id),
			created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate group_lists table: %w", err)
	}

	// 2. Участники
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS group_members (
			id          BIGSERIAL PRIMARY KEY,
			group_id    BIGINT NOT NULL REFERENCES group_lists(id) ON DELETE CASCADE,
			user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			joined_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(group_id, user_id)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate group_members table: %w", err)
	}

	// 3. Приглашения
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS group_invites (
			id          BIGSERIAL PRIMARY KEY,
			group_id    BIGINT NOT NULL REFERENCES group_lists(id) ON DELETE CASCADE,
			inviter_id  BIGINT NOT NULL,
			invitee_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			status      VARCHAR(20) DEFAULT 'pending',
			created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(group_id, invitee_id)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate group_invites table: %w", err)
	}

	// 4. Запись в групповом списке
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS group_items (
			id           BIGSERIAL PRIMARY KEY,
			group_id     BIGINT NOT NULL REFERENCES group_lists(id) ON DELETE CASCADE,
			added_by     BIGINT NOT NULL,
			title        VARCHAR(500) NOT NULL,
			content_type VARCHAR(20) NOT NULL DEFAULT 'movie',
			status       VARCHAR(20) NOT NULL DEFAULT 'planned',
			notes        TEXT NOT NULL DEFAULT '',
			poster_url   TEXT NOT NULL DEFAULT '',
			genres       TEXT[] NOT NULL DEFAULT '{}',
			created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate group_items table: %w", err)
	}

	_, _ = pool.Exec(ctx, `ALTER TABLE group_items ADD COLUMN IF NOT EXISTS current_episode INT NOT NULL DEFAULT 1;`)
	_, _ = pool.Exec(ctx, `ALTER TABLE group_items ADD COLUMN IF NOT EXISTS max_episodes INT NOT NULL DEFAULT 1;`)

	// Миграции для новых полей Anime в группах
	_, _ = pool.Exec(ctx, `ALTER TABLE group_items ADD COLUMN IF NOT EXISTS shikimori_id INTEGER;`)
	_, _ = pool.Exec(ctx, `ALTER TABLE group_items ADD COLUMN IF NOT EXISTS tmdb_id INTEGER;`)
	_, _ = pool.Exec(ctx, `ALTER TABLE group_items ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';`)
	_, _ = pool.Exec(ctx, `ALTER TABLE group_items ADD COLUMN IF NOT EXISTS episodes_total INTEGER;`)
	_, _ = pool.Exec(ctx, `ALTER TABLE group_items ADD COLUMN IF NOT EXISTS aniliberty_alias TEXT NOT NULL DEFAULT '';`)
	_, _ = pool.Exec(ctx, `ALTER TABLE group_items ADD COLUMN IF NOT EXISTS shikimori_score NUMERIC(4,2);`)

	// 5. Персональные оценки участников
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS group_item_ratings (
			id         BIGSERIAL PRIMARY KEY,
			item_id    BIGINT NOT NULL REFERENCES group_items(id) ON DELETE CASCADE,
			user_id    BIGINT NOT NULL,
			rating     SMALLINT CHECK (rating IS NULL OR (rating >= 1 AND rating <= 10)),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(item_id, user_id)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate group_item_ratings table: %w", err)
	}

	// 6. Ссылки к записям
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS group_item_links (
			id         BIGSERIAL PRIMARY KEY,
			item_id    BIGINT NOT NULL REFERENCES group_items(id) ON DELETE CASCADE,
			user_id    BIGINT NOT NULL,
			label      VARCHAR(200) NOT NULL DEFAULT '',
			url        TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate group_item_links table: %w", err)
	}

	// Индексы для групповых таблиц
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_group_members_group_id ON group_members(group_id);`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_group_members_user_id ON group_members(user_id);`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_group_invites_invitee_id ON group_invites(invitee_id);`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_group_items_group_id ON group_items(group_id);`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_group_item_ratings_item_id ON group_item_ratings(item_id);`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_group_item_links_item_id ON group_item_links(item_id);`)

	return pool, nil
}
