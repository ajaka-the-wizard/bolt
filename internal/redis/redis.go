package redis

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/configs"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	rdb *redis.Client
}

func InitializeRedis(ctx context.Context, env *configs.Env, logger *slog.Logger) *Redis {
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
		panic(err)
	}
	logger.Info("Redis cache connected successfully")

	return &Redis{
		rdb,
	}
}
func (r *Redis) CloseConn() error {
	if r.rdb != nil {
		return r.rdb.Close()
	}
	return nil
}
