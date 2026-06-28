package repository

import (
	"context"
	"errors"
	"time"

	"app4every/services/auth/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInviteCodeInvalid = errors.New("invite code is invalid or already used")
)

type InviteRepository interface {
	Create(ctx context.Context, code string, createdBy int64) (*model.InviteCode, error)
	GetByCode(ctx context.Context, code string) (*model.InviteCode, error)
	GetByCreatedBy(ctx context.Context, userID int64) ([]*model.InviteCode, error)
	MarkAsUsed(ctx context.Context, codeID, usedByID int64) error
}

type postgresInviteRepository struct {
	db *pgxpool.Pool
}

func NewInviteRepository(db *pgxpool.Pool) InviteRepository {
	return &postgresInviteRepository{db: db}
}

func (r *postgresInviteRepository) Create(ctx context.Context, code string, createdBy int64) (*model.InviteCode, error) {
	query := `
		INSERT INTO invite_codes (code, created_by)
		VALUES ($1, $2)
		RETURNING id, code, created_by, used_by, created_at, used_at
	`
	invite := &model.InviteCode{}
	err := r.db.QueryRow(ctx, query, code, createdBy).Scan(
		&invite.ID, &invite.Code, &invite.CreatedBy, &invite.UsedBy, &invite.CreatedAt, &invite.UsedAt,
	)
	if err != nil {
		return nil, err
	}
	return invite, nil
}

func (r *postgresInviteRepository) GetByCode(ctx context.Context, code string) (*model.InviteCode, error) {
	query := `
		SELECT id, code, created_by, used_by, created_at, used_at
		FROM invite_codes
		WHERE code = $1
	`
	invite := &model.InviteCode{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&invite.ID, &invite.Code, &invite.CreatedBy, &invite.UsedBy, &invite.CreatedAt, &invite.UsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInviteCodeInvalid
		}
		return nil, err
	}
	return invite, nil
}

func (r *postgresInviteRepository) GetByCreatedBy(ctx context.Context, userID int64) ([]*model.InviteCode, error) {
	query := `
		SELECT id, code, created_by, used_by, created_at, used_at
		FROM invite_codes
		WHERE created_by = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*model.InviteCode
	for rows.Next() {
		invite := &model.InviteCode{}
		if err := rows.Scan(&invite.ID, &invite.Code, &invite.CreatedBy, &invite.UsedBy, &invite.CreatedAt, &invite.UsedAt); err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	if invites == nil {
		invites = []*model.InviteCode{}
	}
	return invites, nil
}

func (r *postgresInviteRepository) MarkAsUsed(ctx context.Context, codeID, usedByID int64) error {
	query := `
		UPDATE invite_codes
		SET used_by = $1, used_at = $2
		WHERE id = $3 AND used_by IS NULL
	`
	tag, err := r.db.Exec(ctx, query, usedByID, time.Now(), codeID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInviteCodeInvalid
	}
	return nil
}
