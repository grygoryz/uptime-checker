package entity

import (
	"database/sql"
	"time"
)

type PingKind string

const (
	PingStart   PingKind = "start"
	PingSuccess PingKind = "success"
	PingFail    PingKind = "fail"
)

type CreatePing struct {
	CheckId   string
	Type      PingKind
	Source    string
	UserAgent string
	Body      string
	Date      time.Time
	Duration  sql.NullInt32
}

type PingTypeAndDate struct {
	Type PingKind  `db:"type"`
	Date time.Time `db:"date"`
}

type GetPingsTotal struct {
	CheckId string
	From    time.Time
	To      time.Time
}

type GetPings struct {
	CheckId string
	From    time.Time
	To      time.Time
	Limit   int
	Offset  int
}

type Ping struct {
	Id        int       `db:"id"`
	Type      PingKind  `db:"type"`
	Source    string    `db:"source"`
	UserAgent string    `db:"user_agent"`
	Body      string    `db:"body"`
	Date      time.Time `db:"date"`
	Duration  *int      `db:"duration"`
}
