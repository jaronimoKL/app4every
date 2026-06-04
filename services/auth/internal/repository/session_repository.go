package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository interface {
	Set(ctx context.Context, token string, userID int64, expires time.Duration) error
	Get(ctx context.Context, token string) (int64, error)
	Delete(ctx context.Context, token string) error
}

type redisSessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) SessionRepository {
	return &redisSessionRepository{rdb: rdb}
}

func (r *redisSessionRepository) Set(ctx context.Context, token string, userID int64, expires time.Duration) error {
	key := fmt.Sprintf("session:%s", token)
	return r.rdb.Set(ctx, key, userID, expires).Err()
}

func (r *redisSessionRepository) Get(ctx context.Context, token string) (int64, error) {
	key := fmt.Sprintf("session:%s", token)
	var userID int64
	err := r.rdb.Get(ctx, key).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *redisSessionRepository) Delete(ctx context.Context, token string) error {
	key := fmt.Sprintf("session:%s", token)
	return r.rdb.Del(ctx, key).Err()
}
