package config

import (
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	Api      Api
	Database Database
	Redis    Redis
}

func New() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	return Config{
		Api:      apiCfg(),
		Database: databaseCfg(),
		Redis:    redisCfg(),
	}
}
