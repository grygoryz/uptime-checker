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
	service := auth.NewService(repository.NewUser(s.db))
	auth.RegisterHandler(s.router, service, s.validator)
}

func (s *Server) initSwagger() {
	s.router.Get("/swagger/*", httpSwagger.Handler())
}
