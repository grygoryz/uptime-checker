package entity

import "time"

type NotificationFlipStatus string

const (
	NotificationFlipUp   NotificationFlipStatus = "up"
	NotificationFlipDown NotificationFlipStatus = "down"
)

type Notification struct {
	CheckName     string
	FlipTo        NotificationFlipStatus
	FlipDate      time.Time
	CheckChannels Channels
}
