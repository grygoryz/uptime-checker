package entity

import (
	"github.com/lib/pq"
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
	Id          string        `db:"id"`
	Name        string        `db:"name"`
	Description string        `db:"description"`
	Interval    int           `db:"interval"`
	Grace       int           `db:"grace"`
	LastPing    *time.Time    `db:"last_ping"`
	NextPing    *time.Time    `db:"next_ping"`
	LastStarted *time.Time    `db:"last_started"`
	Status      CheckStatus   `db:"status"`
	Channels    pq.Int64Array `db:"channels"`
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
	CheckId     string
	UserId      int
	Name        string
	Description string
	Interval    int
	Grace       int
}

type DeleteCheck struct {
	CheckId string
	UserId  int
}

type SetCheckStatus struct {
	CheckId string
	UserId  int
	Status  CheckStatus
}

type AddChannels struct {
	CheckId  string
	Channels []int
}
