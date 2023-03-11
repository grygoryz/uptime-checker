package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gitlab.com/grygoryz/uptime-checker/config"
)

func New(cfg config.Config) *sqlx.DB {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return db
}
