package server

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/auth"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/channel"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
)

func (s *Server) initDomains() {
	session := repository.NewSession(s.redis)
	registry := repository.NewRegistry(s.db)

	s.initAuth(registry, session)
	s.initChannel(registry, session)
	s.initSwagger()
}

func (s *Server) initAuth(registry *repository.Registry, session *repository.Session) {
	service := auth.NewService(registry, session)
	auth.RegisterHandler(s.router, service, s.validator, session)
}

func (s *Server) initChannel(registry *repository.Registry, session *repository.Session) {
	service := channel.NewService(registry, session)
	channel.RegisterHandler(s.router, service, s.validator, session)
}

func (s *Server) initSwagger() {
	s.router.Get("/swagger/*", httpSwagger.Handler())
}
