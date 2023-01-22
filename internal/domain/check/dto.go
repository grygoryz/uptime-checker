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
	Channels    []int64            `json:"channels" validate:"required"`
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

type CreateCheckResponse struct {
	Id string `json:"id" validate:"required"`
}
