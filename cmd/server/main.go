package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "gitlab.com/grygoryz/uptime-checker/docs"
	"gitlab.com/grygoryz/uptime-checker/internal/server"
	"log"
)

// @title Uptime Checker
// @version 0.0.1
func main() {
	log.Println("Starting server...")
	s := server.New()
	s.Init()
	s.Run()
}
