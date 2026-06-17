package redis

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/configs"
	"github.com/ajaka-the-wizard/bolt/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	rdb *redis.Client
}

func InitRedis(ctx context.Context, env *configs.Env, logger *slog.Logger) *Redis {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	rdb := redis.NewClient(&redis.Options{
		Addr:     env.REDIS_ADDR,
		Password: env.REDIS_PASSWORD,
		DB:       0,
		Protocol: 2,
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		logger.Error("Could not ping redis", "error", err)
		rdb.Close()
		panic(err)
	}
	if !env.PRODUCTION {
		err = rdb.FlushAll(ctx).Err()
		if err != nil {
			panic(err)
		}
	}
	logger.Info("Redis cache connected successfully")

	if err := rdb.XGroupCreateMkStream(ctx, domain.BoltRedisInvoiceStreamKey, domain.BoltRedisInvoiceConsumerGroup, "$").Err(); err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logger.Error("Failed to create stream with consumer groups", "error", err.Error())
		panic(err)
	}
	// Ensure stream for invoice processing exists
	ensureStream(ctx, rdb, domain.BoltRedisInvoiceStreamKey, domain.BoltRedisInvoiceConsumerGroup, logger)

	// Ensure stream for webhook processing exists
	ensureStream(ctx, rdb, domain.BoltRedisWebhookStreamKey, domain.BoltRedisWebhookConsumerGroup, logger)

	logger.Info("All consumer stream created successfully")
	return &Redis{
		rdb,
	}
}

func ensureStream(ctx context.Context, rdb *redis.Client, stream string, group string, logger *slog.Logger) {
	if err := rdb.XGroupCreateMkStream(ctx, stream, group, "$").Err(); err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logger.Error("Failed to create stream with consumer groups", "error", err.Error())
		panic(err)
	}
}

func (r *Redis) CloseConn() error {
	if r.rdb != nil {
		return r.rdb.Close()
	}
	return nil
}

func (r *Redis) SetIdemKey(ctx context.Context, key string) error {
	iKey := domain.BoltIdempotencyKey + key
	exp := time.Hour * 24
	err := r.rdb.SetNX(ctx, iKey, key, exp).Err()
	return err
}

func (r *Redis) GetIdemKey(ctx context.Context, key string) (int, error) {
	iKey := domain.BoltIdempotencyKey + key
	val, err := r.rdb.Exists(ctx, iKey).Result()
	if err != nil {
		return 0, err
	}
	return int(val), nil
}
