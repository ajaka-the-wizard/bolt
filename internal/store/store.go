package store

import (
	"context"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/database"
	"github.com/ajaka-the-wizard/bolt/internal/redis"
)

// This is the co-ordinator between redis and database
type Store struct {
	r  *redis.Redis
	db *database.Repo
}

func InitStore(r *redis.Redis, db *database.Repo) *Store {
	return &Store{r, db}
}

// SetIdempotencyKey sets the idempotency key
func (s *Store) SetIdempotencyKey(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := s.r.SetIdemKey(ctx, key)
	return err
}

// CheckKeyExistence checks if a key has been seen or not. it returns a boolean depending
func (s *Store) CheckKeyExistence(ctx context.Context, key string) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	val, err := s.r.GetIdemKey(ctx, key)
	if err != nil || val != 1 {
		return false
	}
	return true
}
