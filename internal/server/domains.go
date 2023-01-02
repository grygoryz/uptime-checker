package server

import (
	httpSwagger "github.com/swaggo/http-swagger"
	_ "gitlab.com/grygoryz/uptime-checker/docs"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/auth"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
)

func (s *Server) initDomains() {
	s.initAuth()
	s.initSwagger()
}

func (s *Server) initAuth() {
	session := repository.NewSession(s.redis)
	service := auth.NewService(repository.NewRegistry(s.db), session)
	auth.RegisterHandler(s.router, service, s.validator, session)
}

func (s *Server) initSwagger() {
	s.router.Get("/swagger/*", httpSwagger.Handler())
}
