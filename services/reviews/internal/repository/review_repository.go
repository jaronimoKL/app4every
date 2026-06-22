package repository

import (
	"context"
	"errors"
	"time"

	"app4every/services/reviews/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrReviewNotFound = errors.New("review not found")
var ErrLinkNotFound = errors.New("link not found")
var ErrGenreNotFound = errors.New("genre not found")

type ReviewRepository interface {
	// Рецензии
	Create(ctx context.Context, userID int64, req model.CreateReviewRequest) (*model.Review, error)
	GetAllByUserID(ctx context.Context, userID int64) ([]*model.Review, error)
	GetByID(ctx context.Context, id, userID int64) (*model.Review, error)
	Update(ctx context.Context, id, userID int64, req model.UpdateReviewRequest) (*model.Review, error)
	Delete(ctx context.Context, id, userID int64) error
	// Ссылки
	AddLink(ctx context.Context, reviewID, userID int64, req model.AddLinkRequest) (*model.ReviewLink, error)
	DeleteLink(ctx context.Context, linkID, reviewID, userID int64) error
	GetLinks(ctx context.Context, reviewID int64) ([]model.ReviewLink, error)
	// Жанры
	AddGenre(ctx context.Context, reviewID, userID int64, name string) (*model.ReviewGenre, error)
	DeleteGenre(ctx context.Context, genreID, reviewID, userID int64) error
	GetGenres(ctx context.Context, reviewID int64) ([]model.ReviewGenre, error)
}

type postgresReviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) ReviewRepository {
	return &postgresReviewRepository{db: db}
}

func (r *postgresReviewRepository) Create(ctx context.Context, userID int64, req model.CreateReviewRequest) (*model.Review, error) {
	now := time.Now()
	review := &model.Review{}
	var rating *int16

	err := r.db.QueryRow(ctx, `
		INSERT INTO reviews (user_id, title, content_type, status, rating, notes, poster_url, shikimori_id, description, episodes_total, aniliberty_alias, shikimori_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, user_id, title, content_type, status, rating, notes, poster_url, shikimori_id, description, episodes_total, aniliberty_alias, shikimori_score, created_at, updated_at
	`, userID, req.Title, req.ContentType, req.Status, req.Rating, req.Notes, req.PosterURL, req.ShikimoriID, req.Description, req.EpisodesTotal, req.AnilibertyAlias, req.ShikimoriScore, now, now).
		Scan(&review.ID, &review.UserID, &review.Title, &review.ContentType, &review.Status,
			&rating, &review.Notes, &review.PosterURL, &review.ShikimoriID, &review.Description, &review.EpisodesTotal, &review.AnilibertyAlias, &review.ShikimoriScore, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		return nil, err
	}
	review.Rating = rating
	review.Links = []model.ReviewLink{}
	review.Genres = []model.ReviewGenre{}
	return review, nil
}

// GetAllByUserID возвращает все рецензии пользователя вместе со ссылками.
// Ссылки загружаются одним дополнительным запросом (IN clause), не N+1.
func (r *postgresReviewRepository) GetAllByUserID(ctx context.Context, userID int64) ([]*model.Review, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, title, content_type, status, rating, notes, poster_url, shikimori_id, description, episodes_total, aniliberty_alias, shikimori_score, created_at, updated_at
		FROM reviews WHERE user_id = $1 ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviewMap := make(map[int64]*model.Review)
	var ids []int64

	for rows.Next() {
		rev := &model.Review{Links: []model.ReviewLink{}, Genres: []model.ReviewGenre{}}
		var rating *int16
		if err := rows.Scan(&rev.ID, &rev.UserID, &rev.Title, &rev.ContentType, &rev.Status,
			&rating, &rev.Notes, &rev.PosterURL, &rev.ShikimoriID, &rev.Description, &rev.EpisodesTotal, &rev.AnilibertyAlias, &rev.ShikimoriScore, &rev.CreatedAt, &rev.UpdatedAt); err != nil {
			return nil, err
		}
		rev.Rating = rating
		reviewMap[rev.ID] = rev
		ids = append(ids, rev.ID)
	}

	if len(ids) == 0 {
		return []*model.Review{}, nil
	}

	// Загружаем все ссылки одним запросом
	linkRows, err := r.db.Query(ctx, `
		SELECT id, review_id, label, url, created_at
		FROM review_links WHERE review_id = ANY($1)
		ORDER BY created_at ASC
	`, ids)
	if err != nil {
		return nil, err
	}
	defer linkRows.Close()

	for linkRows.Next() {
		link := model.ReviewLink{}
		if err := linkRows.Scan(&link.ID, &link.ReviewID, &link.Label, &link.URL, &link.CreatedAt); err != nil {
			return nil, err
		}
		if rev, ok := reviewMap[link.ReviewID]; ok {
			rev.Links = append(rev.Links, link)
		}
	}

	// Загружаем все жанры одним запросом
	genreRows, err := r.db.Query(ctx, `
		SELECT id, review_id, name
		FROM review_genres WHERE review_id = ANY($1)
		ORDER BY id ASC
	`, ids)
	if err != nil {
		return nil, err
	}
	defer genreRows.Close()

	for genreRows.Next() {
		genre := model.ReviewGenre{}
		if err := genreRows.Scan(&genre.ID, &genre.ReviewID, &genre.Name); err != nil {
			return nil, err
		}
		if rev, ok := reviewMap[genre.ReviewID]; ok {
			rev.Genres = append(rev.Genres, genre)
		}
	}

	// Возвращаем в порядке updated_at DESC (порядок ids)
	result := make([]*model.Review, 0, len(ids))
	for _, id := range ids {
		result = append(result, reviewMap[id])
	}
	return result, nil
}

func (r *postgresReviewRepository) GetByID(ctx context.Context, id, userID int64) (*model.Review, error) {
	rev := &model.Review{Links: []model.ReviewLink{}, Genres: []model.ReviewGenre{}}
	var rating *int16
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, title, content_type, status, rating, notes, poster_url, shikimori_id, description, episodes_total, aniliberty_alias, shikimori_score, created_at, updated_at
		FROM reviews WHERE id = $1 AND user_id = $2
	`, id, userID).
		Scan(&rev.ID, &rev.UserID, &rev.Title, &rev.ContentType, &rev.Status,
			&rating, &rev.Notes, &rev.PosterURL, &rev.ShikimoriID, &rev.Description, &rev.EpisodesTotal, &rev.AnilibertyAlias, &rev.ShikimoriScore, &rev.CreatedAt, &rev.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	rev.Rating = rating

	links, err := r.GetLinks(ctx, id)
	if err != nil {
		return nil, err
	}
	rev.Links = links

	genres, err := r.GetGenres(ctx, id)
	if err != nil {
		return nil, err
	}
	rev.Genres = genres

	return rev, nil
}

func (r *postgresReviewRepository) Update(ctx context.Context, id, userID int64, req model.UpdateReviewRequest) (*model.Review, error) {
	now := time.Now()
	rev := &model.Review{Links: []model.ReviewLink{}, Genres: []model.ReviewGenre{}}
	var rating *int16

	err := r.db.QueryRow(ctx, `
		UPDATE reviews
		SET title = $1, content_type = $2, status = $3, rating = $4, notes = $5, poster_url = $6,
		    shikimori_id = $7, description = $8, episodes_total = $9, aniliberty_alias = $10, shikimori_score = $11, updated_at = $12
		WHERE id = $13 AND user_id = $14
		RETURNING id, user_id, title, content_type, status, rating, notes, poster_url, shikimori_id, description, episodes_total, aniliberty_alias, shikimori_score, created_at, updated_at
	`, req.Title, req.ContentType, req.Status, req.Rating, req.Notes, req.PosterURL,
		req.ShikimoriID, req.Description, req.EpisodesTotal, req.AnilibertyAlias, req.ShikimoriScore, now, id, userID).
		Scan(&rev.ID, &rev.UserID, &rev.Title, &rev.ContentType, &rev.Status,
			&rating, &rev.Notes, &rev.PosterURL, &rev.ShikimoriID, &rev.Description, &rev.EpisodesTotal, &rev.AnilibertyAlias, &rev.ShikimoriScore, &rev.CreatedAt, &rev.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	rev.Rating = rating

	links, err := r.GetLinks(ctx, id)
	if err != nil {
		return nil, err
	}
	rev.Links = links

	genres, err := r.GetGenres(ctx, id)
	if err != nil {
		return nil, err
	}
	rev.Genres = genres

	return rev, nil
}

func (r *postgresReviewRepository) Delete(ctx context.Context, id, userID int64) error {
	res, err := r.db.Exec(ctx, `DELETE FROM reviews WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrReviewNotFound
	}
	return nil
}

func (r *postgresReviewRepository) AddLink(ctx context.Context, reviewID, userID int64, req model.AddLinkRequest) (*model.ReviewLink, error) {
	// Проверяем что рецензия принадлежит пользователю
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reviews WHERE id=$1 AND user_id=$2)`, reviewID, userID).Scan(&exists)
	if err != nil || !exists {
		return nil, ErrReviewNotFound
	}

	link := &model.ReviewLink{}
	err = r.db.QueryRow(ctx, `
		INSERT INTO review_links (review_id, label, url) VALUES ($1, $2, $3)
		RETURNING id, review_id, label, url, created_at
	`, reviewID, req.Label, req.URL).
		Scan(&link.ID, &link.ReviewID, &link.Label, &link.URL, &link.CreatedAt)
	return link, err
}

func (r *postgresReviewRepository) DeleteLink(ctx context.Context, linkID, reviewID, userID int64) error {
	// JOIN через reviews чтобы проверить владельца
	res, err := r.db.Exec(ctx, `
		DELETE FROM review_links rl
		USING reviews rv
		WHERE rl.id=$1 AND rl.review_id=$2 AND rv.id=$2 AND rv.user_id=$3
	`, linkID, reviewID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrLinkNotFound
	}
	return nil
}

func (r *postgresReviewRepository) GetLinks(ctx context.Context, reviewID int64) ([]model.ReviewLink, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, review_id, label, url, created_at FROM review_links
		WHERE review_id=$1 ORDER BY created_at ASC
	`, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.ReviewLink
	for rows.Next() {
		var l model.ReviewLink
		if err := rows.Scan(&l.ID, &l.ReviewID, &l.Label, &l.URL, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	if links == nil {
		links = []model.ReviewLink{}
	}
	return links, nil
}

func (r *postgresReviewRepository) AddGenre(ctx context.Context, reviewID, userID int64, name string) (*model.ReviewGenre, error) {
	// Проверяем что рецензия принадлежит пользователю
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reviews WHERE id=$1 AND user_id=$2)`, reviewID, userID).Scan(&exists)
	if err != nil || !exists {
		return nil, ErrReviewNotFound
	}

	genre := &model.ReviewGenre{}
	err = r.db.QueryRow(ctx, `
		INSERT INTO review_genres (review_id, name) VALUES ($1, $2)
		RETURNING id, review_id, name
	`, reviewID, name).Scan(&genre.ID, &genre.ReviewID, &genre.Name)
	return genre, err
}

func (r *postgresReviewRepository) DeleteGenre(ctx context.Context, genreID, reviewID, userID int64) error {
	// JOIN через reviews чтобы проверить владельца
	res, err := r.db.Exec(ctx, `
		DELETE FROM review_genres rg
		USING reviews rv
		WHERE rg.id=$1 AND rg.review_id=$2 AND rv.id=$2 AND rv.user_id=$3
	`, genreID, reviewID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrGenreNotFound
	}
	return nil
}

func (r *postgresReviewRepository) GetGenres(ctx context.Context, reviewID int64) ([]model.ReviewGenre, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, review_id, name FROM review_genres
		WHERE review_id=$1 ORDER BY id ASC
	`, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []model.ReviewGenre
	for rows.Next() {
		var g model.ReviewGenre
		if err := rows.Scan(&g.ID, &g.ReviewID, &g.Name); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	if genres == nil {
		genres = []model.ReviewGenre{}
	}
	return genres, nil
}

