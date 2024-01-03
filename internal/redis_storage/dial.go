package redis_storage

import (
	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/redis/go-redis/v9"
)

func Dial(config *conf.REDISConfiguration) (*redis.Client, error) {
	opt, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	return client, nil
}
