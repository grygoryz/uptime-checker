package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	goredis "github.com/go-redis/redis/v9"
	"github.com/jmoiron/sqlx"
	"gitlab.com/grygoryz/uptime-checker/config"
	"gitlab.com/grygoryz/uptime-checker/internal/middleware"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/logger"
	"gitlab.com/grygoryz/uptime-checker/internal/validate"
	"gitlab.com/grygoryz/uptime-checker/third_party/database"
	"gitlab.com/grygoryz/uptime-checker/third_party/redis"
	"log"
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

func New() *Server {
	return &Server{
		router: chi.NewRouter(),
		cfg:    config.New(),
	}
}

func NewTest() *Server {
	return &Server{
		router: chi.NewRouter(),
		cfg:    config.NewTest(),
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
		log.Printf("Serving at %s\n", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	if err := s.gracefulShutdown(); err != nil {
		log.Fatalf("Server shutdown failed: %+v", err)
	}
	log.Println("Server shutdown success.")
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

	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Api.GracefulTimeout)
	defer func() {
		cancel()
		if err := s.db.Close(); err != nil {
			log.Printf("Database shutdown failed: %+v\n", err)
		}
		log.Println("Database shutdown success.")

		if err := s.redis.Close(); err != nil {
			log.Printf("Redis shutdown failed: %+v\n", err)
		}
		log.Println("Redis shutdown success.")

		// TODO: close rabbit. smtp
	}()

	return s.httpServer.Shutdown(ctx)
}
