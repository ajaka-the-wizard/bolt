package redis

import (
	"context"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (r *Redis) addToQueue(ctx context.Context, id uuid.UUID, stream string) error {
	values := map[string]any{"order_id": id.String(), "max_retries": domain.BoltRedisMaxRetries, "no_of_retries": 0}
	arg := redis.XAddArgs{
		Stream:       stream,
		ID:           "*",
		Values:       values,
		IdempotentID: id.String(),
	}
	return r.rdb.XAdd(ctx, &arg).Err()
}

// AddToInvoiceQueue adds a new message to the invoice generating stream
func (r *Redis) AddToInvoiceQueue(ctx context.Context, id uuid.UUID) error {

	return r.addToQueue(ctx, id, domain.BoltRedisInvoiceStreamKey)
}

// AddToWebhookQueue adds a new message to webhook delivery queue
func (r *Redis) AddToWebhookQueue(ctx context.Context, id uuid.UUID) error {
	return r.addToQueue(ctx, id, domain.BoltRedisWebhookStreamKey)
}

// GetNextOnQueue retrieves the next data from redis streams
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

// This acknowledges the successful completion of a job
func (r *Redis) Ack(ctx context.Context, stream string, group string, ids ...string) error {
	return r.rdb.XAck(ctx, stream, group, ids...).Err()
}
