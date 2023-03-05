package queue

import (
	"context"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"gitlab.com/grygoryz/uptime-checker/config"
	"time"
)

type Queue struct {
	conn            *amqp.Connection
	notifyConnClose chan *amqp.Error

	ch              *amqp.Channel
	notifyChanClose chan *amqp.Error

	done chan struct{}
}

const queueName = "notifications"
const publishTimeout = time.Second * 5
const reconnectDelay = time.Second
const reInitDelay = time.Second

func New(cfg config.Config) *Queue {
	queue := Queue{done: make(chan struct{})}

	addr := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)
	go queue.handleReconnect(addr)

	return &queue
}

// PublishBatch publishes a batch of messages and wait for confirmations from broker
func (q *Queue) PublishBatch(ctx context.Context, messages [][]byte) error {
	confirmations := make([]*amqp.DeferredConfirmation, len(messages))
	for i, m := range messages {
		for {
			confirmation, err := q.ch.PublishWithDeferredConfirmWithContext(
				ctx,
				"",
				queueName,
				false,
				false,
				amqp.Publishing{
					ContentType:  "application/json",
					Body:         m,
					DeliveryMode: amqp.Persistent,
				})
			if err == nil {
				confirmations[i] = confirmation
				break
			}

			select {
			case <-q.done:
				return errors.New("client is shutting down")
			case <-time.After(time.Second):
				log.Err(err).Msg("Retrying publish")
			}
		}

	}

	ctxTimeout, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()

	for _, confirmation := range confirmations {
		acked, err := confirmation.WaitContext(ctxTimeout)
		if err != nil {
			return errors.New("publish timeout exceeded")
		}
		if !acked {
			return fmt.Errorf("publishing nacked with delivery tag: %v", confirmation.DeliveryTag)
		}
	}

	return nil
}

func (q *Queue) Close() error {
	close(q.done)
	err := q.ch.Close()
	if err != nil {
		return err
	}

	err = q.conn.Close()
	if err != nil {
		return err
	}

	return nil
}

// handleReconnect will wait for a connection error on notifyConnClose, and then continuously attempt to reconnect.
func (q *Queue) handleReconnect(addr string) {
	for {
		log.Info().Msg("Attempting to connect")
		conn, err := q.connect(addr)

		if err != nil {
			log.Info().Msg("Failed to connect. Retrying...")
			select {
			case <-q.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		if done := q.handleReInit(conn); done {
			break
		}
	}
}

func (q *Queue) handleReInit(conn *amqp.Connection) bool {
	for {
		err := q.init(conn)
		if err != nil {
			log.Info().Msg("Failed to initialize channel. Retrying...")
			select {
			case <-q.done:
				return true
			case <-time.After(reInitDelay):
			}
			continue
		}

		select {
		case <-q.done:
			return true
		case <-q.notifyConnClose:
			log.Info().Msg("Connection closed. Reconnecting...")
			return false
		case <-q.notifyChanClose:
			log.Info().Msg("Channel closed. Re-running init...")
		}
	}
}

// init will initialize channel & declare queue
func (q *Queue) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.Confirm(false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	q.ch = ch
	q.notifyChanClose = make(chan *amqp.Error, 1)
	q.ch.NotifyClose(q.notifyChanClose)

	return nil
}

// connect will create a new AMQP connection
func (q *Queue) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	q.conn = conn
	q.notifyConnClose = make(chan *amqp.Error, 1)
	q.conn.NotifyClose(q.notifyConnClose)
	log.Info().Msg("Connected!")
	return conn, nil
}
