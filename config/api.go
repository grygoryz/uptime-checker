package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Api struct {
	Host            string        `default:"localhost"`
	Port            string        `required:"true"`
	GracefulTimeout time.Duration `default:"20s" split_words:"true"`
}

func ApiCfg() Api {
	var api Api
	envconfig.MustProcess("API", &api)

	return api
}
