package server

import (
	"gitlab.com/grygoryz/uptime-checker/internal/domain/auth"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
)

func (s *Server) InitDomains() {
	s.InitAuth()
}

func (s *Server) InitAuth() {
	auth.RegisterHandler(s.router, auth.NewService(repository.NewUser(s.db)))
}
