package entity

import (
	"time"
)

type CheckStatus string

const (
	CheckNew     CheckStatus = "new"
	CheckStarted CheckStatus = "started"
	CheckUp      CheckStatus = "up"
	CheckDown    CheckStatus = "down"
	CheckPaused  CheckStatus = "paused"
)

type Check struct {
	Id          string      `db:"id"`
	Name        string      `db:"name"`
	Description string      `db:"description"`
	Interval    int         `db:"interval"`
	Grace       int         `db:"grace"`
	LastPing    *time.Time  `db:"last_ping"`
	NextPing    *time.Time  `db:"next_ping"`
	LastStarted *time.Time  `db:"last_started"`
	Status      CheckStatus `db:"status"`
	Channels    Channels    `db:"channels"`
}

type GetCheck struct {
	Id     string
	UserId int
}

type CreateCheck struct {
	UserId      int
	Name        string
	Description string
	Interval    int
	Grace       int
}

type UpdateCheck struct {
	Id          string
	UserId      int
	Name        string
	Description string
	Interval    int
	Grace       int
}

type DeleteCheck struct {
	Id     string
	UserId int
}

type SetCheckStatus struct {
	Id     string
	UserId int
	Status CheckStatus
}

type AddChannels struct {
	Id       string
	Channels []int
}

type CheckExpired struct {
	Id        string     `db:"id"`
	Name      string     `db:"name"`
	Grace     int        `db:"grace"`
	NextPing  *time.Time `db:"next_ping"`
	UserEmail string     `db:"email"`
	Channels  Channels   `db:"channels"`
}
