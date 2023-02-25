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

func New(testing bool) Config {
	re := regexp.MustCompile(`^(.*` + "uptime-checker" + `)`)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	rootPath := re.Find([]byte(cwd))

	envFile := "/.env"
	if testing {
		envFile += ".test"
	}

	err = godotenv.Load(string(rootPath) + envFile)
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		Api:      apiCfg(),
		Database: databaseCfg(),
		Redis:    redisCfg(),
	}
}
