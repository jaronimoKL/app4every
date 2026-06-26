package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        int64                  `json:"id"`
	UserID    int64                  `json:"user_id"`
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	IsRead    bool                   `json:"is_read"`
	CreatedAt time.Time              `json:"created_at"`
}

type NotificationRepo interface {
	Create(ctx context.Context, n *Notification) error
	GetByUserID(ctx context.Context, userID int64) ([]*Notification, error)
	MarkAsRead(ctx context.Context, userID int64, notificationIDs []int64) error
	Delete(ctx context.Context, userID int64, notificationID int64) error
}

type notificationRepo struct {
	db *pgxpool.Pool
}

func NewNotificationRepo(db *pgxpool.Pool) NotificationRepo {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, n *Notification) error {
	var metaBytes []byte
	if n.Metadata != nil {
		metaBytes, _ = json.Marshal(n.Metadata)
	}

	query := `
		INSERT INTO notifications (user_id, type, message, metadata, is_read, created_at)
		VALUES ($1, $2, $3, $4, false, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, n.UserID, n.Type, n.Message, metaBytes).Scan(&n.ID, &n.CreatedAt)
}

func (r *notificationRepo) GetByUserID(ctx context.Context, userID int64) ([]*Notification, error) {
	query := `
		SELECT id, user_id, type, message, metadata, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		var n Notification
		var metaBytes []byte
		err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Message, &metaBytes, &n.IsRead, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		if metaBytes != nil {
			json.Unmarshal(metaBytes, &n.Metadata)
		} else {
			n.Metadata = make(map[string]interface{})
		}
		notifications = append(notifications, &n)
	}
	return notifications, nil
}

func (r *notificationRepo) MarkAsRead(ctx context.Context, userID int64, notificationIDs []int64) error {
	if len(notificationIDs) == 0 {
		return nil
	}
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE user_id = $1 AND id = ANY($2)
	`
	_, err := r.db.Exec(ctx, query, userID, notificationIDs)
	return err
}

func (r *notificationRepo) Delete(ctx context.Context, userID int64, notificationID int64) error {
	query := `
		DELETE FROM notifications
		WHERE user_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query, userID, notificationID)
	return err
}
