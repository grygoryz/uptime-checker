package entity

import "time"

type Notification struct {
	FlipTo        string
	FlipDate      time.Time
	CheckChannels Channels
	UserEmail     string
}
