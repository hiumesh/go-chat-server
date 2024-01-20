package redis_storage

import (
	"context"

	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/redis/go-redis/v9"
)

func Dial(ctx context.Context, config *conf.REDISConfiguration) (*redis.Client, error) {
	opt, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	_, err = client.Ping(ctx).Result()

	if err != nil {
		return nil, err
	}
	return client, nil
}
