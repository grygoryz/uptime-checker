package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	goredis "github.com/go-redis/redis/v9"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gitlab.com/grygoryz/uptime-checker/config"
	"gitlab.com/grygoryz/uptime-checker/internal/middleware"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/logger"
	"gitlab.com/grygoryz/uptime-checker/internal/validate"
	"gitlab.com/grygoryz/uptime-checker/third_party/database"
	"gitlab.com/grygoryz/uptime-checker/third_party/redis"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	router     *chi.Mux
	httpServer *http.Server

	db    *sqlx.DB
	redis *goredis.Client

	cfg       config.Config
	validator *validate.Validator
}

func New(cfg config.Config) *Server {
	return &Server{
		router: chi.NewRouter(),
		cfg:    cfg,
	}
}

func (s *Server) Init() {
	s.setGlobalMiddleware()
	s.newDatabase()
	s.newRedis()
	s.newValidator()
	s.initDomains()
}

func (s *Server) setGlobalMiddleware() {
	s.router.Use(chiMiddleware.RequestID)

	if s.cfg.Api.DisableLogging == false {
		s.router.Use(logger.Logger())
	}
	s.router.Use(chiMiddleware.Recoverer)
	s.router.Use(middleware.CORS)
}

func (s *Server) newDatabase() {
	s.db = database.New(s.cfg)
}

func (s *Server) newRedis() {
	s.redis = redis.New(s.cfg)
}

func (s *Server) newValidator() {
	s.validator = validate.New()
}

func (s *Server) Run() {
	s.httpServer = &http.Server{
		Addr:    s.cfg.Api.Host + ":" + s.cfg.Api.Port,
		Handler: s.router,
	}

	go func() {
		log.Info().Msgf("Serving at %s\n", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Send()
		}
	}()

	if err := s.gracefulShutdown(); err != nil {
		log.Fatal().Msgf("Server shutdown failed: %+v", err)
	}
	log.Info().Msg("Server shutdown success.")
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) DB() *sqlx.DB {
	return s.db
}

func (s *Server) gracefulShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Info().Msg("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Api.GracefulTimeout)
	defer func() {
		cancel()
		if err := s.db.Close(); err != nil {
			log.Err(err).Msg("Database shutdown failed")
		}
		log.Info().Msg("Database shutdown success.")

		if err := s.redis.Close(); err != nil {
			log.Err(err).Msg("Redis shutdown failed")
		}
		log.Info().Msg("Redis shutdown success.")
	}()

	return s.httpServer.Shutdown(ctx)
}
