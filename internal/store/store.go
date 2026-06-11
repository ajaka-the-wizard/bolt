package store

import (
	"context"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/database"
	"github.com/ajaka-the-wizard/bolt/internal/redis"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type redisClient interface {
	SetIdemKey(ctx context.Context, key string) error
	GetIdemKey(ctx context.Context, key string) (int, error)
	AddToInvoiceQueue(ctx context.Context, id uuid.UUID) error
	GetNextOnQueue(ctx context.Context, id string, stream string, group string) ([]goredis.XStream, error)
}

type Store struct {
	r  redisClient
	db *database.Repo
}


	return &Store{r, db}
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
