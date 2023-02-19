package check

import (
	"github.com/go-chi/chi/v5"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/session"
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

func RegisterHandler(router *chi.Mux, service Service, validator *validate.Validator, sessionRepo *session.Repository) {
	h := handler{service: service, validator: validator}

	authMiddleware := session.Auth(sessionRepo)

	router.Route("/v1/checks", func(router chi.Router) {
		router.Use(authMiddleware)
		router.Get("/", h.GetChecks)
		router.Post("/", h.CreateCheck)
		router.Get("/{id}", h.GetCheck)
		router.Put("/{id}", h.UpdateCheck)
		router.Delete("/{id}", h.DeleteCheck)
		router.Put("/{id}/pause", h.PauseCheck)
		router.Get("/{id}/pings", h.GetPings)
		router.Get("/{id}/flips", h.GetFlips)
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
	user := session.User(r.Context())
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

	user := session.User(r.Context())
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

	user := session.User(r.Context())
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

	respond.JSON(r.Context(), w, http.StatusCreated, CreateCheckResponse{Id: id})
}

// UpdateCheck updates check
// @Tags Checks
// @Summary Update check
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path string true "check id"
// @Param check body UpdateCheckBody true "check data"
// @Success 200
// @router /v1/checks/{id} [put]
func (h handler) UpdateCheck(w http.ResponseWriter, r *http.Request) {
	checkId := chi.URLParam(r, "id")
	err := h.validator.Struct(CheckIdParam{Id: checkId})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	body, err := request.Body[UpdateCheckBody](r, h.validator)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := session.User(r.Context())
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

	user := session.User(r.Context())
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

	user := session.User(r.Context())
	err = h.service.PauseCheck(r.Context(), checkId, user.Id)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusOK)
}

// GetPings returns check's pings
// @Tags Checks
// @Summary Get pings
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path string true "check id"
// @Param params query GetPingsQuery true "params"
// @Success 200 {object} GetPingsResponse
// @router /v1/checks/{id}/pings [get]
func (h handler) GetPings(w http.ResponseWriter, r *http.Request) {
	checkId := chi.URLParam(r, "id")
	err := h.validator.Struct(CheckIdParam{Id: checkId})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	limit, err := request.IntQueryParam(r, "limit")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}
	offset, err := request.IntQueryParam(r, "offset")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}
	from, err := request.IntQueryParam(r, "from")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}
	to, err := request.IntQueryParam(r, "to")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	err = h.validator.Struct(GetPingsQuery{
		From:   from,
		To:     to,
		Limit:  limit,
		Offset: &offset,
	})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	pings, total, err := h.service.GetPings(r.Context(), entity.GetPings{
		CheckId: checkId,
		From:    time.UnixMilli(int64(from)),
		To:      time.UnixMilli(int64(to)),
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	items := make([]Ping, len(pings))
	for i, ping := range pings {
		items[i] = Ping{
			Id:        ping.Id,
			Type:      ping.Type,
			Source:    ping.Source,
			UserAgent: ping.UserAgent,
			Body:      ping.Body,
			Date:      ping.Date.UTC(),
			Duration:  ping.Duration,
		}
	}

	respond.JSON(r.Context(), w, http.StatusOK, GetPingsResponse{
		Total: total,
		Items: items,
	})
}

// GetFlips returns check's flips
// @Tags Checks
// @Summary Get flips
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path string true "check id"
// @Param params query GetFlipsQuery true "params"
// @Success 200 {object} GetFlipsResponse
// @router /v1/checks/{id}/flips [get]
func (h handler) GetFlips(w http.ResponseWriter, r *http.Request) {
	checkId := chi.URLParam(r, "id")
	err := h.validator.Struct(CheckIdParam{Id: checkId})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	limit, err := request.IntQueryParam(r, "limit")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}
	offset, err := request.IntQueryParam(r, "offset")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}
	from, err := request.IntQueryParam(r, "from")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}
	to, err := request.IntQueryParam(r, "to")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	err = h.validator.Struct(GetFlipsQuery{
		From:   from,
		To:     to,
		Limit:  limit,
		Offset: &offset,
	})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	flips, total, err := h.service.GetFlips(r.Context(), entity.GetFlips{
		CheckId: checkId,
		From:    time.UnixMilli(int64(from)),
		To:      time.UnixMilli(int64(to)),
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	items := make([]Flip, len(flips))
	for i, flip := range flips {
		items[i] = Flip{
			To:   flip.To,
			Date: flip.Date.UTC(),
		}
	}

	respond.JSON(r.Context(), w, http.StatusOK, GetFlipsResponse{
		Total: total,
		Items: items,
	})
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
