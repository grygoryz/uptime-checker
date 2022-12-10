package main

import (
	"gitlab.com/grygoryz/uptime-checker/internal/server"
	"log"
)

func main() {
	log.Println("Starting server...")
	s := server.New()
	s.Init()
	s.Run()
}
