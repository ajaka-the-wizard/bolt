package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Store) FetchNextTask(ctx context.Context, id string, stream string, group string, logger *slog.Logger) ([]redis.XStream, error) {
	var data []redis.XStream
	var err error
	backoff := time.Second
	maxBackoff := 30 * time.Second
	for {
		data, err = s.r.GetNextOnQueue(ctx, id, stream, group)
		if err != nil {
			logger.Error("Failed to fetch next task", "error", err.Error())
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff = backoff * 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}
		break
	}
	return data, nil
}
