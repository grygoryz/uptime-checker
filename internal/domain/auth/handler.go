package auth

import (
	"github.com/go-chi/chi/v5"
	"gitlab.com/grygoryz/uptime-checker/internal/middleware"
	"gitlab.com/grygoryz/uptime-checker/internal/respond"
	"net/http"
)

type handler struct {
	service Service
}

func RegisterHandler(router *chi.Mux, service Service) {
	h := handler{service: service}

	router.Route("/v1/auth", func(router chi.Router) {
		router.Post("/signin", h.SignIn)
	})
}

func (h handler) SignIn(w http.ResponseWriter, r *http.Request) {
	log := middleware.LogEntry(r.Context())
	log.Info().Msg("Hello!")
	_, err := h.service.SignIn()
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}
	respond.Status(w, http.StatusOK)
}
