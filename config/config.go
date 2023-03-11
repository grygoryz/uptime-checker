package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
)

type Config struct {
	Api      Api
	Database Database
	Redis    Redis
	RabbitMQ RabbitMQ
	Mailjet  Mailjet
}

func New(testing bool) Config {
	re := regexp.MustCompile(`^(.*` + "uptime-checker" + `)`)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	rootPath := re.Find([]byte(cwd))

	envFile := "/.env"
	if testing {
		envFile += ".test"
	}

	err = godotenv.Load(string(rootPath) + envFile)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return Config{
		Api:      apiCfg(),
		Database: databaseCfg(),
		Redis:    redisCfg(),
		RabbitMQ: rabbitMQCfg(),
		Mailjet:  mailjetCfg(),
	}
}
