package redis

import (
	"context"
	"im-system/pkg/config"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() error {
	Client = redis.NewClient(&redis.Options{
		Addr:     config.Global.Redis.Addr,
		Password: config.Global.Redis.Password,
		DB:       config.Global.Redis.DB,
	})
	return Client.Ping(context.Background()).Err()
}
