package repository

import (
	"context"
	"errors"
	"time"

	"app4every/services/notebook/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoteNotFound = errors.New("note not found")

type NoteRepository interface {
	Create(ctx context.Context, userID int64, title, content string) (*model.Note, error)
	GetByUserID(ctx context.Context, userID int64) ([]*model.Note, error)
	GetByID(ctx context.Context, id, userID int64) (*model.Note, error)
	Update(ctx context.Context, id, userID int64, title, content string) (*model.Note, error)
	Delete(ctx context.Context, id, userID int64) error
}

type postgresNoteRepository struct {
	db *pgxpool.Pool
}

func NewNoteRepository(db *pgxpool.Pool) NoteRepository {
	return &postgresNoteRepository{db: db}
}

func (r *postgresNoteRepository) Create(ctx context.Context, userID int64, title, content string) (*model.Note, error) {
	query := `
		INSERT INTO notes (user_id, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, content, created_at, updated_at
	`
	now := time.Now()
	note := &model.Note{}
	err := r.db.QueryRow(ctx, query, userID, title, content, now, now).Scan(
		&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return note, nil
}

func (r *postgresNoteRepository) GetByUserID(ctx context.Context, userID int64) ([]*model.Note, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM notes
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*model.Note
	for rows.Next() {
		note := &model.Note{}
		if err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	// Вернуть пустой слайс вместо nil если нет заметок
	if notes == nil {
		notes = []*model.Note{}
	}
	return notes, nil
}

// GetByID проверяет что заметка принадлежит указанному userID — защита от IDOR.
func (r *postgresNoteRepository) GetByID(ctx context.Context, id, userID int64) (*model.Note, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM notes WHERE id = $1 AND user_id = $2
	`
	note := &model.Note{}
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoteNotFound
		}
		return nil, err
	}
	return note, nil
}

// Update также фильтрует по userID чтобы пользователь не мог изменить чужую заметку.
func (r *postgresNoteRepository) Update(ctx context.Context, id, userID int64, title, content string) (*model.Note, error) {
	query := `
		UPDATE notes SET title = $1, content = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5
		RETURNING id, user_id, title, content, created_at, updated_at
	`
	note := &model.Note{}
	err := r.db.QueryRow(ctx, query, title, content, time.Now(), id, userID).Scan(
		&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoteNotFound
		}
		return nil, err
	}
	return note, nil
}

func (r *postgresNoteRepository) Delete(ctx context.Context, id, userID int64) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM notes WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNoteNotFound
	}
	return nil
}
