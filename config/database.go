package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Database struct {
	Host     string `required:"true"`
	Port     string `required:"true"`
	User     string `required:"true"`
	Password string `required:"true"`
	Name     string `required:"true"`
}

func databaseCfg() Database {
	var db Database
	envconfig.MustProcess("DB", &db)

	return db
}
