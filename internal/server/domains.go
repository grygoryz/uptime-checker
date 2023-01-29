package server

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/auth"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/channel"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/check"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/ping"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/session"
)

func (s *Server) initDomains() {
	sessionRepo := session.NewRepository(s.redis)
	registry := repository.NewRegistry(s.db)

	s.initAuth(registry, sessionRepo)
	s.initChannel(registry, sessionRepo)
	s.initCheck(registry, sessionRepo)
	s.initPing(registry)
	s.initSwagger()
}

func (s *Server) initAuth(registry *repository.Registry, sessionRepo *session.Repository) {
	service := auth.NewService(registry, sessionRepo)
	auth.RegisterHandler(s.router, service, s.validator, sessionRepo)
}

func (s *Server) initChannel(registry *repository.Registry, sessionRepo *session.Repository) {
	service := channel.NewService(registry)
	channel.RegisterHandler(s.router, service, s.validator, sessionRepo)
}

func (s *Server) initCheck(registry *repository.Registry, sessionRepo *session.Repository) {
	service := check.NewService(registry)
	check.RegisterHandler(s.router, service, s.validator, sessionRepo)
}

func (s *Server) initPing(registry *repository.Registry) {
	service := ping.NewService(registry)
	ping.RegisterHandler(s.router, service, s.validator)
}

func (s *Server) initSwagger() {
	s.router.Get("/swagger/*", httpSwagger.Handler())
}
