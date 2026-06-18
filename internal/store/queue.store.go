package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// FetchNextTask is responsible for fetching the next message from the stream, it continues to retry until a message is received
func (s *Store) FetchNextTask(ctx context.Context, id string, stream string, group string, logger *slog.Logger) ([]redis.XStream, error) {
	var data []redis.XStream
	var err error
	backoff := time.Second
	maxBackoff := 30 * time.Second
	for {
		data, err = s.r.GetNextOnQueue(ctx, id, stream, group)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			logger.Error("Failed to fetch next task", "error", err.Error())
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff = backoff * 2
			backoff = min(backoff, maxBackoff)
			continue
		}
		break
	}
	return data, nil
}

// Ack acknowledge a certain stream message has been processed
func (s *Store) Ack(ctx context.Context, stream string, group string, id ...string) error {
	return s.r.Ack(ctx, stream, group, id...)
}
