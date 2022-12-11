package server

import (
	"gitlab.com/grygoryz/uptime-checker/internal/domain/auth"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
)

func (s *Server) initDomains() {
	s.initAuth()
}

func (s *Server) initAuth() {
	service := auth.NewService(repository.NewUser(s.db))
	auth.RegisterHandler(s.router, service, s.validator)
}
