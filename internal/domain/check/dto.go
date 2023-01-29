package check

import (
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"time"
)

type Check struct {
	Id          string             `json:"id" validate:"required"`
	Name        string             `json:"name" validate:"required"`
	Description string             `json:"description" validate:"required"`
	Interval    int                `json:"interval" validate:"required"`
	Grace       int                `json:"grace" validate:"required"`
	LastPing    *time.Time         `json:"lastPing,omitempty"`
	NextPing    *time.Time         `json:"nextPing,omitempty"`
	LastStarted *time.Time         `json:"lastStarted,omitempty"`
	Status      entity.CheckStatus `json:"status" validate:"required"`
	Channels    []Channel          `json:"channels" validate:"required"`
}

type Channel struct {
	Id         int                `json:"id" validate:"required"`
	Kind       entity.ChannelKind `json:"kind" validate:"required"`
	Email      *string            `json:"email,omitempty"`
	WebhookURL *string            `json:"webhookURL,omitempty"`
}

type CheckIdParam struct {
	Id string `json:"id" validate:"uuid4"`
}

type CreateCheckBody struct {
	Name        string `json:"name" validate:"required,max=128"`
	Description string `json:"description" validate:"required,max=528"`
	Interval    int    `json:"interval" validate:"required,min=60,max=31536000"` // min 1 minute, max 1 year
	Grace       int    `json:"grace" validate:"required,min=60,max=31536000"`    // min 1 minute, max 1 year
	Channels    []int  `json:"channels" validate:"required,min=1"`
}

type UpdateCheckBody struct {
	CreateCheckBody
}

type CreateCheckResponse struct {
	Id string `json:"id" validate:"required"`
}

type GetPingsQuery struct {
	From   int  `json:"from" validate:"required"`
	To     int  `json:"to" validate:"required"`
	Limit  int  `json:"limit" validate:"required,min=1,max=50"`
	Offset *int `json:"offset" validate:"required"`
}

type Ping struct {
	Id        int             `json:"id" validate:"required"`
	Type      entity.PingKind `json:"type" validate:"required"`
	Source    string          `json:"source" validate:"required"`
	UserAgent string          `json:"userAgent" validate:"required"`
	Body      string          `json:"body,omitempty"`
	Date      time.Time       `json:"date" validate:"required"`
	Duration  *int            `json:"duration,omitempty"`
}

type GetPingsResponse struct {
	Total int    `json:"total" validate:"required"`
	Items []Ping `json:"items" validate:"required"`
}

type GetFlipsQuery struct {
	From   int  `json:"from" validate:"required"`
	To     int  `json:"to" validate:"required"`
	Limit  int  `json:"limit" validate:"required,min=1,max=50"`
	Offset *int `json:"offset" validate:"required"`
}

type Flip struct {
	To   entity.FlipState `json:"to" validate:"required"`
	Date time.Time        `json:"date" validate:"required"`
}

type GetFlipsResponse struct {
	Total int    `json:"total" validate:"required"`
	Items []Flip `json:"items" validate:"required"`
}
