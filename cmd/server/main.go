package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
	"gitlab.com/grygoryz/uptime-checker/config"
	_ "gitlab.com/grygoryz/uptime-checker/docs"
	"gitlab.com/grygoryz/uptime-checker/internal/server"
)

// @title Uptime Checker
// @version 0.0.1
// @securitydefinitions.apikey cookieAuth
// @in                         cookie
// @name                       sessionId
func main() {
	log.Info().Msg("Starting server...")
	cfg := config.New(false)
	s := server.New(cfg)
	s.Init()
	s.Run()
}
