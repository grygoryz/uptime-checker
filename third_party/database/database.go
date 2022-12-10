package database

import (
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"gitlab.com/grygoryz/uptime-checker/config"
	"log"
)

func New(cfg config.Config) *sqlx.DB {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	return db
}
