package entity

type ChannelKind string

const (
	EmailChannel   ChannelKind = "email"
	WebhookChannel ChannelKind = "webhook"
)

type Channel struct {
	Id         int    `db:"id"`
	Kind       string `db:"kind"`
	Email      string `db:"email"`
	WebhookURL string `db:"webhook_url"`
	UserId     int    `db:"user_id"`
}
