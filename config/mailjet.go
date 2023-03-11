package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Mailjet struct {
	ApiKey      string `required:"true" split_words:"true"`
	SecretKey   string `required:"true" split_words:"true"`
	SenderName  string `required:"true" split_words:"true"`
	SenderEmail string `required:"true" split_words:"true"`
}

func mailjetCfg() Mailjet {
	var mailjet Mailjet
	envconfig.MustProcess("MAILJET", &mailjet)

	return mailjet
}
