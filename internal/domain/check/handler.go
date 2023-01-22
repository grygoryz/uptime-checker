package check

import (
	"github.com/go-chi/chi/v5"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/middleware"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/request"
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

	router.Route("/v1/checks", func(router chi.Router) {
		router.Use(authMiddleware)
		router.Get("/", h.GetChecks)
		router.Post("/", h.CreateCheck)
		router.Get("/{id}", h.GetCheck)
		router.Put("/{id}", h.UpdateCheck)
		router.Delete("/{id}", h.DeleteCheck)
		router.Put("/{id}/pause", h.PauseCheck)
		router.Put("/{id}/resume", h.ResumeCheck)
	})
}

// GetChecks returns user's checks
// @Tags Checks
// @Summary Get checks
// @Security cookieAuth
// @Accept json
// @Produce json
// @Success 200 {array} Check
// @router /v1/checks [get]
func (h handler) GetChecks(w http.ResponseWriter, r *http.Request) {
	user := middleware.User(r.Context())
	checks, err := h.service.GetChecks(r.Context(), user.Id)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	response := make([]Check, len(checks))
	for i, check := range checks {
		response[i] = checkDTO(check)
	}
	respond.JSON(r.Context(), w, http.StatusOK, response)
}

// GetCheck returns user's check by id
// @Tags Checks
// @Summary Get check
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path string true "check id"
// @Success 200 {object} Check
// @router /v1/checks/{id} [get]
func (h handler) GetCheck(w http.ResponseWriter, r *http.Request) {
	checkId := chi.URLParam(r, "id")
	err := h.validator.Struct(CheckIdParam{Id: checkId})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := middleware.User(r.Context())
	check, err := h.service.GetCheck(r.Context(), user.Id, checkId)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.JSON(r.Context(), w, http.StatusOK, checkDTO(check))
}

// CreateCheck creates check and returns its id
// @Tags Checks
// @Summary Create check
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param check body CreateCheckBody true "check data"
// @Success 201 {object} CreateCheckResponse
// @router /v1/checks [post]
func (h handler) CreateCheck(w http.ResponseWriter, r *http.Request) {
	body, err := request.Body[CreateCheckBody](r, h.validator)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := middleware.User(r.Context())
	id, err := h.service.CreateCheck(r.Context(), entity.CreateCheck{
		UserId:      user.Id,
		Name:        body.Name,
		Description: body.Description,
		Interval:    body.Interval,
		Grace:       body.Grace,
	}, body.Channels)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.JSON(r.Context(), w, http.StatusOK, CreateCheckResponse{Id: id})
}

// UpdateCheck updates check
// @Tags Checks
// @Summary Update check
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path string true "check id"
// @Param check body CreateCheckBody true "check data"
// @Success 200
// @router /v1/checks/{id} [put]
func (h handler) UpdateCheck(w http.ResponseWriter, r *http.Request) {
	checkId := chi.URLParam(r, "id")
	err := h.validator.Struct(CheckIdParam{Id: checkId})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	body, err := request.Body[CreateCheckBody](r, h.validator)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := middleware.User(r.Context())
	err = h.service.UpdateCheck(r.Context(), entity.UpdateCheck{
		Id:          checkId,
		UserId:      user.Id,
		Name:        body.Name,
		Description: body.Description,
		Interval:    body.Interval,
		Grace:       body.Grace,
	}, body.Channels)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusOK)
}

// DeleteCheck deletes check
// @Tags Checks
// @Summary Delete check
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path string true "check id"
// @Success 200
// @router /v1/checks/{id} [delete]
func (h handler) DeleteCheck(w http.ResponseWriter, r *http.Request) {
	checkId := chi.URLParam(r, "id")
	err := h.validator.Struct(CheckIdParam{Id: checkId})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := middleware.User(r.Context())
	err = h.service.DeleteCheck(r.Context(), entity.DeleteCheck{
		Id:     checkId,
		UserId: user.Id,
	})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusOK)
}

// PauseCheck pauses check
// @Tags Checks
// @Summary Pause check
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path string true "check id"
// @Success 200
// @router /v1/checks/{id}/pause [put]
func (h handler) PauseCheck(w http.ResponseWriter, r *http.Request) {
	checkId := chi.URLParam(r, "id")
	err := h.validator.Struct(CheckIdParam{Id: checkId})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := middleware.User(r.Context())
	err = h.service.PauseCheck(r.Context(), checkId, user.Id)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusOK)
}

// ResumeCheck resumes check
// @Tags Checks
// @Summary Pause check
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path string true "check id"
// @Success 200
// @router /v1/checks/{id}/resume [put]
func (h handler) ResumeCheck(w http.ResponseWriter, r *http.Request) {
	checkId := chi.URLParam(r, "id")
	err := h.validator.Struct(CheckIdParam{Id: checkId})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := middleware.User(r.Context())
	err = h.service.ResumeCheck(r.Context(), checkId, user.Id)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusOK)
}

// checkDTO transforms entity.Check to Check
func checkDTO(check entity.Check) Check {
	response := Check{
		Id:          check.Id,
		Name:        check.Name,
		Description: check.Description,
		Interval:    check.Interval,
		Grace:       check.Grace,
		Status:      check.Status,
		LastPing:    utc(check.LastPing),
		NextPing:    utc(check.NextPing),
		LastStarted: utc(check.LastStarted),
		Channels:    make([]Channel, len(check.Channels)),
	}
	for i, channel := range check.Channels {
		response.Channels[i] = Channel{
			Id:         channel.Id,
			Kind:       channel.Kind,
			Email:      channel.Email,
			WebhookURL: channel.WebhookURL,
		}
	}

	return response
}

func utc(time *time.Time) *time.Time {
	if time == nil {
		return nil
	}
	val := time.UTC()
	return &val
}
