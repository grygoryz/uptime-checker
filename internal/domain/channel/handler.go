package channel

import (
	"github.com/go-chi/chi/v5"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/session"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/request"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/respond"
	"gitlab.com/grygoryz/uptime-checker/internal/validate"
	"net/http"
)

type handler struct {
	service   *service
	validator *validate.Validator
}

func RegisterHandler(router *chi.Mux, service *service, validator *validate.Validator, sessionRepo *session.Repository) {
	h := handler{service: service, validator: validator}

	authMiddleware := session.Auth(sessionRepo)

	router.Route("/v1/channels", func(router chi.Router) {
		router.Use(authMiddleware)
		router.Get("/", h.GetChannels)
		router.Post("/", h.CreateChannel)
		router.Put("/{id}", h.UpdateChannel)
		router.Delete("/{id}", h.DeleteChannel)
	})
}

// CreateChannel creates channel
// @Tags Channels
// @Summary Create channel
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param channel body CreateChannelBody true "channel data"
// @Success 201 {object} CreateChannelResponse
// @router /v1/channels [post]
func (h handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	body, err := request.Body[CreateChannelBody](r, h.validator)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := session.User(r.Context())
	id, err := h.service.CreateChannel(r.Context(), entity.CreateChannel{
		Kind:       body.Kind,
		Email:      body.Email,
		WebhookURL: body.WebhookURL,
		UserId:     user.Id,
	})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.JSON(r.Context(), w, http.StatusCreated, CreateChannelResponse{Id: id})
}

// UpdateChannel updates channel
// @Tags Channels
// @Summary Update channel
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path int true "channel id"
// @Param channel body UpdateChannelBody true "channel data"
// @Success 200
// @router /v1/channels/{id} [put]
func (h handler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	body, err := request.Body[UpdateChannelBody](r, h.validator)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	id, err := request.IntParam(r, "id")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := session.User(r.Context())
	err = h.service.UpdateChannel(r.Context(), entity.Channel{
		Id:         id,
		Kind:       body.Kind,
		Email:      body.Email,
		WebhookURL: body.WebhookURL,
		UserId:     user.Id,
	})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusOK)
}

// GetChannels returns channels
// @Tags Channels
// @Summary Get channels
// @Security cookieAuth
// @Accept json
// @Produce json
// @Success 200 {array} GetChannelsResponseItem
// @router /v1/channels [get]
func (h handler) GetChannels(w http.ResponseWriter, r *http.Request) {
	user := session.User(r.Context())
	channels, err := h.service.GetChannels(r.Context(), user.Id)
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	response := make([]GetChannelsResponseItem, len(channels))
	for i, channel := range channels {
		response[i] = GetChannelsResponseItem{
			Id:         channel.Id,
			Kind:       channel.Kind,
			Email:      channel.Email,
			WebhookURL: channel.WebhookURL,
		}
	}
	respond.JSON(r.Context(), w, http.StatusOK, response)
}

// DeleteChannel deletes channel by id
// @Tags Channels
// @Summary Delete channel
// @Security cookieAuth
// @Accept json
// @Produce json
// @Param id path int true "channel id"
// @Success 200
// @router /v1/channels/{id} [delete]
func (h handler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	id, err := request.IntParam(r, "id")
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	user := session.User(r.Context())
	err = h.service.DeleteChannel(r.Context(), entity.DeleteChannel{Id: id, UserId: user.Id})
	if err != nil {
		respond.Error(r.Context(), w, err)
		return
	}

	respond.Status(w, http.StatusOK)
}
