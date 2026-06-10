package redis

import (
	"context"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (r *Redis) AddToReportGenQueue(ctx context.Context, id uuid.UUID) error {
	arg := redis.XAddArgs{Stream: domain.BoltRedisStreamKey, ID: "*", Values: id.String(), IdempotentID: id.String()}
	return r.rdb.XAdd(ctx, &arg).Err()
}
