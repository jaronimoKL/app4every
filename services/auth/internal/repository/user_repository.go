package repository

import (
	"context"
	"errors"
	"time"

	"app4every/services/auth/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository interface {
	Create(ctx context.Context, username, email, passwordHash string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByIdentifier(ctx context.Context, identifier string) (*model.User, error) // email ИЛИ username
	GetByID(ctx context.Context, id int64) (*model.User, error)
	UpdateProfile(ctx context.Context, id int64, username, email string) (*model.User, error)
	UpdatePassword(ctx context.Context, id int64, newPasswordHash string) error
}

type postgresUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{db: db}
}

func isPgUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// scanUser читает строку из БД в структуру User.
// email nullable — если NULL, оставляем пустую строку.
func scanUser(row pgx.Row, user *model.User) error {
	var emailPtr *string
	err := row.Scan(
		&user.ID,
		&user.Username,
		&emailPtr,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if emailPtr != nil {
		user.Email = *emailPtr
	}
	return nil
}

// scanUserNoHash — для запросов где password_hash не нужен в ответе (Create, UpdateProfile).
func scanUserNoHash(row pgx.Row, user *model.User) error {
	var emailPtr *string
	err := row.Scan(
		&user.ID,
		&user.Username,
		&emailPtr,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if emailPtr != nil {
		user.Email = *emailPtr
	}
	return nil
}

func (r *postgresUserRepository) Create(ctx context.Context, username, email, passwordHash string) (*model.User, error) {
	// Если email пустой — храним NULL (UNIQUE разрешает несколько NULL)
	var emailArg interface{}
	if email != "" {
		emailArg = email
	}

	query := `
		INSERT INTO users (username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, created_at, updated_at
	`
	now := time.Now()
	user := &model.User{}
	row := r.db.QueryRow(ctx, query, username, emailArg, passwordHash, now, now)
	if err := scanUserNoHash(row, user); err != nil {
		if isPgUniqueViolation(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`
	user := &model.User{}
	if err := scanUser(r.db.QueryRow(ctx, query, email), user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetByIdentifier ищет пользователя по email ИЛИ username.
// Сервер сам определяет что имел ввиду пользователь — клиент шлёт одно поле.
func (r *postgresUserRepository) GetByIdentifier(ctx context.Context, identifier string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1 OR username = $1
		LIMIT 1
	`
	user := &model.User{}
	if err := scanUser(r.db.QueryRow(ctx, query, identifier), user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users WHERE id = $1
	`
	user := &model.User{}
	if err := scanUser(r.db.QueryRow(ctx, query, id), user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) UpdateProfile(ctx context.Context, id int64, username, email string) (*model.User, error) {
	var emailArg interface{}
	if email != "" {
		emailArg = email
	}

	query := `
		UPDATE users SET username = $1, email = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, username, email, created_at, updated_at
	`
	user := &model.User{}
	row := r.db.QueryRow(ctx, query, username, emailArg, time.Now(), id)
	if err := scanUserNoHash(row, user); err != nil {
		if isPgUniqueViolation(err) {
			return nil, ErrUserAlreadyExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) UpdatePassword(ctx context.Context, id int64, newPasswordHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`,
		newPasswordHash, time.Now(), id,
	)
	return err
}
