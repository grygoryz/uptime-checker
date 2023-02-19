package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
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

func NewTest() Config {
	re := regexp.MustCompile(`^(.*` + "uptime-checker" + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))
	err := godotenv.Load(string(rootPath) + `/.env.test`)
	if err != nil {
		log.Println(err)
	}

	return Config{
		Api:      apiCfg(),
		Database: databaseCfg(),
		Redis:    redisCfg(),
	}
}
