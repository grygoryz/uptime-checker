package auth

import (
	"github.com/go-chi/chi/v5"
	"gitlab.com/grygoryz/uptime-checker/internal/session"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/request"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/respond"
	"gitlab.com/grygoryz/uptime-checker/internal/validate"
	"net/http"
	"time"
)

type handler struct {
	service   *service
	validator *validate.Validator
}

func RegisterHandler(router *chi.Mux, service *service, validator *validate.Validator, sessionRepo *session.Repository) {
	h := handler{service: service, validator: validator}

	authMiddleware := session.Auth(sessionRepo)

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
// @Param user body SignUpBody true "user data"
// @Success 201
// @router /v1/auth/signup [put]
func (h handler) SignUp(w http.ResponseWriter, r *http.Request) {
	body, err := request.Body[SignUpBody](r, h.validator)
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
// @Description Sets cookie "sessionId" on response
// @Accept json
// @Produce json
// @Param credentials body SignInBody true "user credentials"
// @Success 200
// @router /v1/auth/signin [put]
func (h handler) SignIn(w http.ResponseWriter, r *http.Request) {
	body, err := request.Body[SignInBody](r, h.validator)
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
		Name:     session.CookieName,
		Value:    id,
		Expires:  time.Now().Add(session.TTL),
		HttpOnly: true,
		Path:     "/",
	})

	respond.Status(w, http.StatusOK)
}

// SignOut destroys user's session
// @Tags Auth
// @Summary Sign out
// @Security cookieAuth
// @Accept json
// @Produce json
// @Success 200
// @router /v1/auth/signout [put]
func (h handler) SignOut(w http.ResponseWriter, r *http.Request) {
	user := session.User(r.Context())

	err := h.service.SignOut(r.Context(), user.SessionId)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	// delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	})

	respond.Status(w, http.StatusOK)
}

// Check returns user session data
// @Tags Auth
// @Summary Check user data
// @Security cookieAuth
// @Accept json
// @Produce json
// @Success 200 {object} CheckResponse
// @router /v1/auth/check [get]
func (h handler) Check(w http.ResponseWriter, r *http.Request) {
	user := session.User(r.Context())
	respond.JSON(r.Context(), w, http.StatusOK, CheckResponse{Id: user.Id, Email: user.Email})
}
