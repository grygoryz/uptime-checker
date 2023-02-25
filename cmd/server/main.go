package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"gitlab.com/grygoryz/uptime-checker/config"
	_ "gitlab.com/grygoryz/uptime-checker/docs"
	"gitlab.com/grygoryz/uptime-checker/internal/server"
	"log"
)

// @title Uptime Checker
// @version 0.0.1
// @securitydefinitions.apikey cookieAuth
// @in                         cookie
// @name                       sessionId
func main() {
	log.Println("Starting server...")
	cfg := config.New(false)
	s := server.New(cfg)
	s.Init()
	s.Run()
}
