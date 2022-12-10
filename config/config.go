package config

import (
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	Api      Api
	Database Database
}

func New() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	return Config{
		Api:      ApiCfg(),
		Database: DatabaseCfg(),
	}
}
