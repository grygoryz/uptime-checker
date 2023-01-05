package entity

import "database/sql"

type ChannelKind string

const (
	EmailChannel   ChannelKind = "email"
	WebhookChannel ChannelKind = "webhook"
)

type Channel struct {
	Id         int         `db:"id"`
	Kind       ChannelKind `db:"kind"`
	Email      string      `db:"email"`
	WebhookURL string      `db:"webhook_url"`
	UserId     int         `db:"user_id"`
}

type ChannelShort struct {
	Id         int            `db:"id"`
	Kind       ChannelKind    `db:"kind"`
	Email      sql.NullString `db:"email"`
	WebhookURL sql.NullString `db:"webhook_url"`
}

type CreateChannel struct {
	Kind       ChannelKind
	Email      string
	WebhookURL string
	UserId     int
}

type DeleteChannel struct {
	Id     int
	UserId int
}
