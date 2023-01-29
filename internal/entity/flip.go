package entity

import "time"

type FlipState string

const (
	FlipUp     FlipState = "up"
	FlipDown   FlipState = "down"
	FlipPaused FlipState = "paused"
)

type CreateFlip struct {
	To      FlipState
	Date    time.Time
	CheckId string
}

type GetFlipsTotal struct {
	CheckId string
	From    time.Time
	To      time.Time
}

type GetFlips struct {
	CheckId string
	From    time.Time
	To      time.Time
	Limit   int
	Offset  int
}

type Flip struct {
	To   FlipState `db:"to"`
	Date time.Time `db:"date"`
}
