package store

import (
	"context"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/database"
	"github.com/ajaka-the-wizard/bolt/internal/queues"
	"github.com/ajaka-the-wizard/bolt/internal/redis"
)

type Store struct {
	r  *redis.Redis
	db *database.Repo
	q  *queues.Queue
}

func InitStore(r *redis.Redis, db *database.Repo, q *queues.Queue) *Store {
	return &Store{r, db, q}
}

func (s *Store) SetIdempotencyKey(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := s.r.SetIdemKey(ctx, key)
	return err
}
func (s *Store) CheckKeyExistence(ctx context.Context, key string) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	val, err := s.r.GetIdemKey(ctx, key)
	if err != nil || val != 1 {
		return false
	}
	return true
}
