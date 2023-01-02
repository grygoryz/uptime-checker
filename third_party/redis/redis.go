package redis

import (
	"context"
	"github.com/go-redis/redis/v9"
	"gitlab.com/grygoryz/uptime-checker/config"
	"log"
)

func New(cfg config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Host + ":" + cfg.Redis.Port})

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalln(err)
	}

	return rdb
}
