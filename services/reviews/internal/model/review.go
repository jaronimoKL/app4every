package model

import "time"

// ── Типы контента ──

type ContentType string

const (
	TypeMovie  ContentType = "movie"
	TypeAnime  ContentType = "anime"
	TypeSeries ContentType = "series"
)

// ── Статусы просмотра ──

type ReviewStatus string

const (
	StatusWatching  ReviewStatus = "watching"
	StatusCompleted ReviewStatus = "completed"
	StatusPlanned   ReviewStatus = "planned"
	StatusDropped   ReviewStatus = "dropped"
	StatusOnHold    ReviewStatus = "on_hold"
)

// ── Сущности ──

// ReviewGenre — жанр рецензии.
type ReviewGenre struct {
	ID       int64  `json:"id"`
	ReviewID int64  `json:"review_id"`
	Name     string `json:"name"`
}

// ReviewLink — ссылка прикреплённая к рецензии пользователем.
// label: "Kinopoisk", "IMDB", "Shikimori" и т.д.
type ReviewLink struct {
	ID        int64     `json:"id"`
	ReviewID  int64     `json:"review_id"`
	Label     string    `json:"label"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type Review struct {
	ID              int64        `json:"id"`
	UserID          int64        `json:"user_id"`
	Title           string       `json:"title"`
	ContentType     ContentType  `json:"content_type"`
	Status          ReviewStatus `json:"status"`
	Rating          *int16       `json:"rating"`    // nullable: nil = без оценки
	Notes           string       `json:"notes"`
	PosterURL       string       `json:"poster_url"` // URL постера (необязателен)
	ShikimoriID     *int         `json:"shikimori_id,omitempty"`
	Description     string       `json:"description"`
	EpisodesTotal   *int         `json:"episodes_total,omitempty"`
	CurrentEpisode  int          `json:"current_episode"`
	AnilibertyAlias string       `json:"aniliberty_alias"`
	ShikimoriScore  *float64     `json:"shikimori_score,omitempty"`
	Links           []ReviewLink `json:"links"`
	Genres          []ReviewGenre `json:"genres"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// ── Request-типы ──

type CreateReviewRequest struct {
	Title           string       `json:"title"`
	ContentType     ContentType  `json:"content_type"`
	Status          ReviewStatus `json:"status"`
	Rating          *int16       `json:"rating"`
	Notes           string       `json:"notes"`
	PosterURL       string       `json:"poster_url"`
	ShikimoriID     *int         `json:"shikimori_id,omitempty"`
	Description     string       `json:"description"`
	EpisodesTotal   *int         `json:"episodes_total,omitempty"`
	CurrentEpisode  int          `json:"current_episode"`
	AnilibertyAlias string       `json:"aniliberty_alias"`
	ShikimoriScore  *float64     `json:"shikimori_score,omitempty"`
}

type UpdateReviewRequest struct {
	Title           string       `json:"title"`
	ContentType     ContentType  `json:"content_type"`
	Status          ReviewStatus `json:"status"`
	Rating          *int16       `json:"rating"`
	Notes           string       `json:"notes"`
	PosterURL       string       `json:"poster_url"`
	ShikimoriID     *int         `json:"shikimori_id,omitempty"`
	Description     string       `json:"description"`
	EpisodesTotal   *int         `json:"episodes_total,omitempty"`
	CurrentEpisode  int          `json:"current_episode"`
	AnilibertyAlias string       `json:"aniliberty_alias"`
	ShikimoriScore  *float64     `json:"shikimori_score,omitempty"`
}

type AddLinkRequest struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}
