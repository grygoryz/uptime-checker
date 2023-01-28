package ping

import (
	"github.com/go-chi/chi/v5"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/logger"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/respond"
	"gitlab.com/grygoryz/uptime-checker/internal/validate"
	"io"
	"net/http"
	"time"
)

type handler struct {
	service   Service
	validator *validate.Validator
}

func RegisterHandler(router *chi.Mux, service Service, validator *validate.Validator) {
	h := handler{service: service, validator: validator}

	router.Route("/v1/pings", func(router chi.Router) {
		router.Put("/{checkId}", h.CreateSuccessPing)
		router.Put("/{checkId}/start", h.CreateStartPing)
		router.Put("/{checkId}/fail", h.CreateFailPing)
	})
}

// maxBodySize defines max size of body in bytes
const maxBodySize = 300000

// CreateSuccessPing creates success ping
// @Tags Pings
// @Summary Create success ping
// @Accept json
// @Produce json
// @Param body body string false "body"
// @Param checkId path string true "check id"
// @Success 200
// @router /v1/pings/{checkId} [put]
func (h handler) CreateSuccessPing(w http.ResponseWriter, r *http.Request) {
	h.pingHandler(w, r, entity.PingSuccess)
}

// CreateStartPing creates start ping
// @Tags Pings
// @Summary Create start ping
// @Accept json
// @Produce json
// @Param body body string false "body"
// @Param checkId path string true "check id"
// @Success 200
// @router /v1/pings/{checkId}/start [put]
func (h handler) CreateStartPing(w http.ResponseWriter, r *http.Request) {
	h.pingHandler(w, r, entity.PingStart)
}

// CreateFailPing creates fail ping
// @Tags Pings
// @Summary Create fail ping
// @Accept json
// @Produce json
// @Param body body string false "body"
// @Param checkId path string true "check id"
// @Success 200
// @router /v1/pings/{checkId}/fail [put]
func (h handler) CreateFailPing(w http.ResponseWriter, r *http.Request) {
	h.pingHandler(w, r, entity.PingFail)
}

func (h handler) pingHandler(w http.ResponseWriter, r *http.Request, kind entity.PingKind) {
	log := logger.LogEntry(r.Context())

	checkId := chi.URLParam(r, "checkId")
	err := h.validator.Struct(CheckIdParam{Id: checkId})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	body, err := io.ReadAll(r.Body)
	if err != nil && err != io.EOF {
		log.Error().Err(err).Send()
	}

	err = h.service.CreatePing(r.Context(), entity.CreatePing{
		CheckId:   checkId,
		Type:      kind,
		Source:    r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Body:      string(body),
		Date:      time.Now(),
	})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusOK)
}
