package redis

import (
	"context"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (r *Redis) AddToInvoiceQueue(ctx context.Context, id uuid.UUID) error {
	arg := redis.XAddArgs{Stream: domain.BoltRedisInvoiceStreamKey, ID: "*", Values: map[string]any{"order_id": id.String(), "max_retries": domain.BoltRedisMaxRetries, "no_of_retries": 0}, IdempotentID: id.String()}
	return r.rdb.XAdd(ctx, &arg).Err()
}
func (r *Redis) GetNextOnQueue(ctx context.Context, id string, stream string, group string) ([]redis.XStream, error) {
	a := redis.XReadGroupArgs{
		Group:    group,
		Consumer: id,
		Streams:  []string{stream, ">"},
		Count:    1,
		Block:    0,
		NoAck:    false,
	}
	data, err := r.rdb.XReadGroup(ctx, &a).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}
