package repository

import (
	"context"
	"errors"

	"app4every/services/auth/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTokenNotFound = errors.New("token not found or expired")
)

type VerificationTokenRepository interface {
	Create(ctx context.Context, userID int64, token, purpose string, expiresAtUnix int64) (*model.VerificationToken, error)
	GetByToken(ctx context.Context, token, purpose string) (*model.VerificationToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUserIDAndPurpose(ctx context.Context, userID int64, purpose string) error
}

type postgresVerificationTokenRepository struct {
	db *pgxpool.Pool
}

func NewVerificationTokenRepository(db *pgxpool.Pool) VerificationTokenRepository {
	return &postgresVerificationTokenRepository{db: db}
}

func (r *postgresVerificationTokenRepository) Create(ctx context.Context, userID int64, token, purpose string, expiresAtUnix int64) (*model.VerificationToken, error) {
	query := `
		INSERT INTO verification_tokens (user_id, token, purpose, expires_at)
		VALUES ($1, $2, $3, to_timestamp($4))
		RETURNING id, user_id, token, purpose, expires_at, created_at
	`
	t := &model.VerificationToken{}
	err := r.db.QueryRow(ctx, query, userID, token, purpose, expiresAtUnix).Scan(
		&t.ID, &t.UserID, &t.Token, &t.Purpose, &t.ExpiresAt, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *postgresVerificationTokenRepository) GetByToken(ctx context.Context, token, purpose string) (*model.VerificationToken, error) {
	query := `
		SELECT id, user_id, token, purpose, expires_at, created_at
		FROM verification_tokens
		WHERE token = $1 AND purpose = $2 AND expires_at > CURRENT_TIMESTAMP
	`
	t := &model.VerificationToken{}
	err := r.db.QueryRow(ctx, query, token, purpose).Scan(
		&t.ID, &t.UserID, &t.Token, &t.Purpose, &t.ExpiresAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *postgresVerificationTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM verification_tokens WHERE token = $1", token)
	return err
}

func (r *postgresVerificationTokenRepository) DeleteByUserIDAndPurpose(ctx context.Context, userID int64, purpose string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM verification_tokens WHERE user_id = $1 AND purpose = $2", userID, purpose)
	return err
}
