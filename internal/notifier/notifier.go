// Package notifier implements a notification system that consumes flip notifications from a queue concurrently
// and notifies users through provided channels. The notifier handles emails and webhooks, and sends
// them to the respective recipients.
package notifier

import (
	"encoding/json"
	"fmt"
	"github.com/mailjet/mailjet-apiv3-go"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/grygoryz/uptime-checker/config"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/queue"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type notifier struct {
	q   *queue.Queue
	mj  *mailjet.Client
	cfg config.Config
}

var client = http.Client{
	Timeout: time.Second * 5,
}

const concurrency = 100

func New(cfg config.Config) *notifier {
	q := queue.New(cfg)
	mj := mailjet.NewMailjetClient(cfg.Mailjet.ApiKey, cfg.Mailjet.SecretKey)
	mj.SetClient(&client)

	return &notifier{
		q:   q,
		mj:  mj,
		cfg: cfg,
	}
}

func (n *notifier) Start() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	err := n.q.Consume(n.handler, concurrency)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	<-quit
	n.shutdown()
}

func (n *notifier) handler(msg amqp091.Delivery) bool {
	log := log.With().Str("messageId", msg.MessageId).Logger()

	log.Info().Msgf("Processing notification: %v", string(msg.Body))
	var notification entity.Notification
	err := json.Unmarshal(msg.Body, &notification)
	if err != nil {
		log.Err(err).Msgf("Unmarshalling message failed with id %v", msg.MessageId)
		return false
	}

	// extract emails and webhooks
	var emails []string
	var webhooks []string
	for _, channel := range notification.CheckChannels {
		switch channel.Kind {
		case entity.EmailChannel:
			emails = append(emails, *channel.Email)
		case entity.WebhookChannel:
			if notification.FlipTo == entity.NotificationFlipUp {
				webhooks = append(webhooks, *channel.WebhookURLUp)
			} else {
				webhooks = append(webhooks, *channel.WebhookURLDown)
			}
		}
	}

	// send emails and trigger webhooks concurrently
	var wg sync.WaitGroup
	wg.Add(len(emails) + len(webhooks))

	go func() {
		n.sendEmail(&log, emails, notification.CheckName, notification.FlipTo, notification.FlipDate)
		wg.Done()
	}()

	for _, webhook := range webhooks {
		go func(webhook string) {
			n.triggerWebhook(&log, webhook)
			wg.Done()
		}(webhook)
	}
	wg.Wait()

	return true
}

func (n *notifier) sendEmail(log *zerolog.Logger, to []string, checkName string, status entity.NotificationFlipStatus, date time.Time) {
	message := mailjet.InfoMessagesV31{
		From: &mailjet.RecipientV31{
			Email: n.cfg.Mailjet.SenderEmail,
			Name:  n.cfg.Mailjet.SenderName,
		},
	}

	recipients := make(mailjet.RecipientsV31, len(to))
	for i, email := range to {
		recipients[i] = mailjet.RecipientV31{Email: email}
	}
	message.To = &recipients

	switch status {
	case entity.NotificationFlipDown:
		message.Subject = fmt.Sprintf("Check %v is down", checkName)
		message.TextPart = fmt.Sprintf("Your check %v is down. Date: %v", checkName, date.UTC().String())
	case entity.NotificationFlipUp:
		message.Subject = fmt.Sprintf("Check %v is up", checkName)
		message.TextPart = fmt.Sprintf("Your check %v is up. Date: %v", checkName, date.UTC().String())
	}

	messages := mailjet.MessagesV31{Info: []mailjet.InfoMessagesV31{message}}
	_, err := n.mj.SendMailV31(&messages)
	if err != nil {
		log.Err(err).Msg("Send email failed")
	}
	log.Info().Msg("Send email success")
}

func (n *notifier) triggerWebhook(log *zerolog.Logger, webhook string) {
	var err error
	var res *http.Response
	retries := 3
	for {
		if retries == 0 {
			break
		}
		res, err = client.Get(webhook)
		if err != nil {
			log.Err(err).Msg("Request failed with error, retrying...")
			retries--
			continue
		}
		if res.StatusCode != http.StatusOK {
			log.Error().Msg("Request got non-200 status code, retrying...")
			retries--
			continue
		}
		break
	}
	log.Info().Msgf("Webhook trigger success: %v", webhook)
}

func (n *notifier) shutdown() {
	log.Info().Msg("Shutting down...")
	if err := n.q.Close(); err != nil {
		log.Info().Msgf("Queue shutdown failed: %+v", err)
	}
	log.Info().Msg("Shutdown success.")
}
