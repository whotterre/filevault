package repositories

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type SessionRepository interface {
	CheckSessionExists(ctx context.Context, email string) (bool, error)
	CreateSession(ctx context.Context, email, sessionID string, ttl time.Duration) error
	DeleteSession(ctx context.Context, sessionID string) (bool, error)
}

type sessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(client *redis.Client) SessionRepository {
	return &sessionRepository{
		client: client,
	}
}

func (r *sessionRepository) CheckSessionExists(ctx context.Context, email string) (bool, error) {
	exists, err := r.client.Exists(ctx, email).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *sessionRepository) CreateSession(ctx context.Context, email, sessionID string, exp time.Duration) error {
	err := r.client.Set(ctx, email, sessionID, exp).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *sessionRepository) DeleteSession(ctx context.Context, sessionID string) (bool, error) {
	delCount, err := r.client.Del(ctx, sessionID).Result()
	if err != nil {
		return false, err
	}
	return delCount > 0, nil
}