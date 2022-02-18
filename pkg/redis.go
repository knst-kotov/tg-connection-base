package pkg

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type redisDB struct {
	Client *redis.Client
}

func NewRedisClient(addr string, ctx context.Context) (*redisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "NewDatabase ping")
	}
	return &redisDB{Client: client}, nil
}
