package main

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/grygoryz/uptime-checker/config"
	"gitlab.com/grygoryz/uptime-checker/internal/notifier"
)

func main() {
	log.Info().Msg("Starting notifier...")
	cfg := config.New(false)
	p := notifier.New(cfg)
	p.Start()
}
