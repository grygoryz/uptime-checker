package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
	"gitlab.com/grygoryz/uptime-checker/config"
	"gitlab.com/grygoryz/uptime-checker/internal/poller"
)

func main() {
	log.Info().Msg("Starting poller...")
	cfg := config.New(false)
	p := poller.New(cfg)
	p.Start()
}
