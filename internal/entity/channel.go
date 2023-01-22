package entity

import (
	"encoding/json"
	"fmt"
)

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
	Id         int         `db:"id"`
	Kind       ChannelKind `db:"kind"`
	Email      *string     `db:"email"`
	WebhookURL *string     `db:"webhook_url"`
}

type Channels []ChannelShort

// Scan converts the data returned from the DB into the struct.
func (c *Channels) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, c)
	case string:
		return json.Unmarshal([]byte(vv), c)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
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
