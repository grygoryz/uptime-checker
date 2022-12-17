package main

import (
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
