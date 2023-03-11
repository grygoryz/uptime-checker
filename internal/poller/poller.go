package poller

import (
	"context"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gitlab.com/grygoryz/uptime-checker/config"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/queue"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/third_party/database"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type poller struct {
	db *sqlx.DB
	r  *repository.Registry
	q  *queue.Queue
}

func New(cfg config.Config) *poller {
	db := database.New(cfg)
	q := queue.New(cfg)

	return &poller{
		db: db,
		q:  q,
		r:  repository.NewRegistry(db),
	}
}

func (p *poller) Start() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case <-quit:
			p.shutdown()
			break loop
		default:
			p.poll()
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (p *poller) poll() {
	log.Info().Msg("Poll start")
	err := p.r.WithTx(context.Background(), func(ctx context.Context) error {
		// get expired checks
		expired, err := p.r.Check.GetExpired(ctx)
		if err != nil {
			return err
		}
		log.Info().Msgf("Expired checks: %+v", expired)

		// update expired checks status to down
		if len(expired) > 0 {
			checkIds := make([]string, len(expired))
			for i, check := range expired {
				checkIds[i] = check.Id
			}
			err = p.r.Check.SetDown(ctx, checkIds)
			if err != nil {
				return err
			}
		}

		// get unprocessed flips
		flips, err := p.r.Flip.GetUnprocessed(ctx)
		if err != nil {
			return err
		}
		log.Info().Msgf("Unprocessed flips: %+v", flips)

		// create new flips from expired checks
		var newFlipsIds []int
		var newFlips []entity.CreateFlip
		if len(expired) > 0 {
			newFlips = make([]entity.CreateFlip, len(expired))
			for i, check := range expired {
				newFlips[i] = entity.CreateFlip{
					To:      entity.FlipDown,
					Date:    check.NextPing.Add(time.Second * time.Duration(check.Grace)),
					CheckId: check.Id,
				}
			}
			newFlipsIds, err = p.r.Flip.CreateMany(ctx, newFlips)
			if err != nil {
				return err
			}
			log.Info().Msgf("Created flips ids: %v", newFlipsIds)
		}

		if len(expired) == 0 && len(flips) == 0 {
			log.Info().Msg("No flips to process")
			return nil
		}

		// build messages and send to queue
		err = p.sendToQueue(ctx, expired, newFlips, flips)
		if err != nil {
			return err
		}

		// set processed: true to all flips
		flipIds := make([]int, 0, len(newFlipsIds)+len(flips))
		for _, id := range newFlipsIds {
			flipIds = append(flipIds, id)
		}
		for _, flip := range flips {
			flipIds = append(flipIds, flip.Id)
		}
		err = p.r.Flip.SetProcessed(ctx, flipIds)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Err(err).Msg("Polling transaction error")
	}
	log.Info().Msg("Poll end")
}

func (p *poller) sendToQueue(
	ctx context.Context,
	expired []entity.CheckExpired,
	newFlips []entity.CreateFlip,
	flips []entity.FlipUnprocessed,
) error {
	notifications := make([][]byte, 0, len(expired)+len(flips))
	for _, flip := range flips {
		n := entity.Notification{
			CheckName:     flip.CheckName,
			FlipTo:        entity.NotificationFlipStatus(flip.To),
			FlipDate:      flip.Date,
			CheckChannels: flip.CheckChannels,
		}
		j, err := json.Marshal(n)
		if err != nil {
			return err
		}
		notifications = append(notifications, j)
	}
	for i, check := range expired {
		n := entity.Notification{
			CheckName:     check.Name,
			FlipTo:        entity.NotificationFlipStatus(newFlips[i].To),
			FlipDate:      newFlips[i].Date,
			CheckChannels: check.Channels,
		}
		j, err := json.Marshal(n)
		if err != nil {
			return err
		}
		notifications = append(notifications, j)
	}

	return p.q.PublishBatch(ctx, notifications)
}

func (p *poller) shutdown() {
	log.Info().Msg("Shutting down...")
	if err := p.db.Close(); err != nil {
		log.Info().Msgf("Database shutdown failed: %+v", err)
	}
	if err := p.q.Close(); err != nil {
		log.Info().Msgf("Queue shutdown failed: %+v", err)
	}
	log.Info().Msg("Database shutdown success.")
	log.Info().Msg("Shutdown success.")
}
