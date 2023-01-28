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
	Type PingKind
	Date time.Time
}
