package auth

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"gitlab.com/grygoryz/uptime-checker/internal/middleware"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/respond"
	"gitlab.com/grygoryz/uptime-checker/internal/validate"
	"net/http"
	"time"
)

type handler struct {
	service   Service
	validator *validate.Validator
}

func RegisterHandler(router *chi.Mux, service Service, validator *validate.Validator, session *repository.Session) {
	h := handler{service: service, validator: validator}

	authMiddleware := middleware.Auth(session)

	router.Route("/v1/auth", func(router chi.Router) {
		router.Put("/signin", h.SignIn)
		router.Put("/signup", h.SignUp)
		router.With(authMiddleware).Put("/signout", h.SignOut)
		router.With(authMiddleware).Get("/check", h.Check)
	})
}

// SignUp creates user
// @Tags Auth
// @Summary Sign up
// @Accept json
// @Produce json
// @Param User body SignUpBody true "user data"
// @Success 201
// @router /v1/auth/signup [put]
func (h handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var body SignUpBody
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

	err = h.service.SignUp(r.Context(), body)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusCreated)
}

// SignIn checks user's credentials and sends session id in cookie
// @Tags Auth
// @Summary Sign in
// @Accept json
// @Produce json
// @Param Credentials body SignInBody true "user credentials"
// @Success 200
// @router /v1/auth/signin [put]
func (h handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var body SignInBody
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

	id, err := h.service.SignIn(r.Context(), body)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    id,
		Expires:  time.Now().Add(168 * time.Hour),
		HttpOnly: true,
	})

	respond.Status(w, http.StatusOK)
}

// SignOut destroys user's session
// @Tags Auth
// @Summary Sign out
// @Accept json
// @Produce json
// @Success 200
// @router /v1/auth/signout [put]
func (h handler) SignOut(w http.ResponseWriter, r *http.Request) {
	user := middleware.User(r.Context())

	err := h.service.SignOut(r.Context(), user.SessionId)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	// delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	})

	respond.Status(w, http.StatusOK)
}

// Check returns user session data
// @Tags Auth
// @Summary Check user data
// @Accept json
// @Produce json
// @Success 200 {object} CheckResponse
// @router /v1/auth/check [get]
func (h handler) Check(w http.ResponseWriter, r *http.Request) {
	user := middleware.User(r.Context())
	respond.JSON(r.Context(), w, http.StatusOK, CheckResponse{Id: user.Id, Email: user.Email})
}
