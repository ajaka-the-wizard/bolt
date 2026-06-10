package store

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func (s *Store) FetchNextTask(id string, stream string, group string) []redis.XStream {
	ctx := context.TODO()
	var data []redis.XStream
	var err error
	for {
		data, err = s.r.GetNextOnQueue(ctx, id, stream, group)
		if err != nil {
			continue
		} else {
			break
		}
	}
	return data
}
