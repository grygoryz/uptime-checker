package auth

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"gitlab.com/grygoryz/uptime-checker/internal/middleware"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/respond"
	"gitlab.com/grygoryz/uptime-checker/internal/validate"
	"net/http"
)

type handler struct {
	service   Service
	validator *validate.Validator
}

func RegisterHandler(router *chi.Mux, service Service, validator *validate.Validator) {
	h := handler{service: service, validator: validator}

	router.Route("/v1/auth", func(router chi.Router) {
		router.Post("/signin", h.SignIn)
		router.Post("/signup", h.CreateUser)
	})
}

// SignIn checks user's credentials and sends api key
// @Tags Auth
// @Summary Sign in
// @Accept json
// @Produce json
// @Success 200
// @router /v1/auth/signin [put]
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

// CreateUser creates user
// @Tags Auth
// @Summary Sign up
// @Accept json
// @Produce json
// @Param User body CreateUserBody true "user data"
// @Success 201
// @router /v1/auth/signup [put]
func (h handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body CreateUserBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	err = h.validator.Struct(body)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	err = h.validator.Struct(CreateUserParams{"John"})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusCreated)
}
