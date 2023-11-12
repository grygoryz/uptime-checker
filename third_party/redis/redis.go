package redis

import (
	"context"
	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog/log"
	"gitlab.com/grygoryz/uptime-checker/config"
)

func New(cfg config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Host + ":" + cfg.Redis.Port})

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return rdb
}
