package config

import (
	"github.com/kelseyhightower/envconfig"
)

type RabbitMQ struct {
	Host     string `required:"true"`
	Port     string `required:"true"`
	User     string `required:"true"`
	Password string `required:"true"`
}

func rabbitMQCfg() RabbitMQ {
	var rabbitMQ RabbitMQ
	envconfig.MustProcess("RABBITMQ", &rabbitMQ)

	return rabbitMQ
}
