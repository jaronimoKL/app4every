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

	// Дружба
	GetFriendship(ctx context.Context, userID, targetID int64) (*model.Friendship, error)
	GetFriends(ctx context.Context, userID int64) ([]*model.User, error)
	GetRequests(ctx context.Context, userID int64) ([]*model.User, error)
	CreateFriendship(ctx context.Context, userID, targetID int64, status string) error
	UpdateFriendshipStatus(ctx context.Context, userID, targetID int64, status string) error
	DeleteFriendship(ctx context.Context, userID, targetID int64) error
	SearchUsers(ctx context.Context, q string, excludeID int64) ([]*model.User, error)
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

func (r *postgresUserRepository) GetFriendship(ctx context.Context, userID, targetID int64) (*model.Friendship, error) {
	query := `
		SELECT id, user_id, friend_id, status, created_at
		FROM friendships
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
	`
	f := &model.Friendship{}
	err := r.db.QueryRow(ctx, query, userID, targetID).Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

func (r *postgresUserRepository) GetFriends(ctx context.Context, userID int64) ([]*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.updated_at
		FROM users u
		JOIN friendships f ON (f.user_id = $1 AND f.friend_id = u.id) OR (f.friend_id = $1 AND f.user_id = u.id)
		WHERE f.status = 'accepted'
		ORDER BY u.username ASC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []*model.User
	for rows.Next() {
		u := &model.User{}
		var emailPtr *string
		if err := rows.Scan(&u.ID, &u.Username, &emailPtr, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if emailPtr != nil {
			u.Email = *emailPtr
		}
		friends = append(friends, u)
	}
	if friends == nil {
		friends = []*model.User{}
	}
	return friends, nil
}

func (r *postgresUserRepository) GetRequests(ctx context.Context, userID int64) ([]*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.updated_at
		FROM users u
		JOIN friendships f ON f.user_id = u.id
		WHERE f.friend_id = $1 AND f.status = 'pending'
		ORDER BY f.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.User
	for rows.Next() {
		u := &model.User{}
		var emailPtr *string
		if err := rows.Scan(&u.ID, &u.Username, &emailPtr, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if emailPtr != nil {
			u.Email = *emailPtr
		}
		requests = append(requests, u)
	}
	if requests == nil {
		requests = []*model.User{}
	}
	return requests, nil
}

func (r *postgresUserRepository) CreateFriendship(ctx context.Context, userID, targetID int64, status string) error {
	query := `
		INSERT INTO friendships (user_id, friend_id, status, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := r.db.Exec(ctx, query, userID, targetID, status)
	return err
}

func (r *postgresUserRepository) UpdateFriendshipStatus(ctx context.Context, userID, targetID int64, status string) error {
	query := `
		UPDATE friendships SET status = $3
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
	`
	_, err := r.db.Exec(ctx, query, userID, targetID, status)
	return err
}

func (r *postgresUserRepository) DeleteFriendship(ctx context.Context, userID, targetID int64) error {
	query := `
		DELETE FROM friendships
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
	`
	res, err := r.db.Exec(ctx, query, userID, targetID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("friendship not found")
	}
	return nil
}

func (r *postgresUserRepository) SearchUsers(ctx context.Context, q string, excludeID int64) ([]*model.User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE username ILIKE $1 AND id <> $2
		ORDER BY username ASC
		LIMIT 20
	`
	rows, err := r.db.Query(ctx, query, "%"+q+"%", excludeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var emailPtr *string
		if err := rows.Scan(&u.ID, &u.Username, &emailPtr, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if emailPtr != nil {
			u.Email = *emailPtr
		}
		users = append(users, u)
	}
	if users == nil {
		users = []*model.User{}
	}
	return users, nil
}
