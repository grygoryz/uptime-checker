package channel

import "gitlab.com/grygoryz/uptime-checker/internal/entity"

type CreateChannelBody struct {
	Kind           entity.ChannelKind `json:"kind" validate:"required,oneof=email webhook"`
	Email          string             `json:"email" validate:"required_if=Kind email,omitempty,email"`
	WebhookURLUp   string             `json:"webhookURLUp" validate:"required_if=Kind webhook,omitempty"`
	WebhookURLDown string             `json:"webhookURLDown" validate:"required_if=Kind webhook,omitempty"`
}

type CreateChannelResponse struct {
	Id int `json:"id" validate:"required"`
}

type UpdateChannelBody struct {
	Kind           entity.ChannelKind `json:"kind" validate:"required,oneof=email webhook"`
	Email          string             `json:"email" validate:"required_if=Kind email,omitempty,email"`
	WebhookURLUp   string             `json:"webhookURLUp" validate:"required_if=Kind webhook,omitempty"`
	WebhookURLDown string             `json:"webhookURLDown" validate:"required_if=Kind webhook,omitempty"`
}

type GetChannelsResponseItem struct {
	Id             int                `json:"id" validate:"required"`
	Kind           entity.ChannelKind `json:"kind" validate:"required"`
	Email          *string            `json:"email,omitempty"`
	WebhookURLUp   *string            `json:"webhookURLUp,omitempty"`
	WebhookURLDown *string            `json:"webhookURLDown,omitempty"`
}
