package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Redis struct {
	Host string `required:"true"`
	Port string `required:"true"`
}

func RedisCfg() Redis {
	var redis Redis
	envconfig.MustProcess("REDIS", &redis)

	return redis
}
