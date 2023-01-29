package ping

import (
	"context"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"math"
)

type Service interface {
	CreatePing(ctx context.Context, ping entity.CreatePing) error
}

type service struct {
	r *repository.Registry
}

func NewService(repositoryRegistry *repository.Registry) Service {
	return &service{r: repositoryRegistry}
}

func (s *service) CreatePing(ctx context.Context, ping entity.CreatePing) error {
	return s.r.WithTx(ctx, func(ctx context.Context) error {
		status, err := s.r.Check.GetStatus(ctx, ping.CheckId)
		if err != nil {
			return err
		}

		switch ping.Type {
		case entity.PingStart:
			err = s.r.Check.PingStart(ctx, ping.CheckId, ping.Date)
		case entity.PingSuccess:
			err = s.r.Check.PingSuccess(ctx, ping.CheckId, ping.Date)
			if err != nil {
				return err
			}

			if status != entity.CheckUp {
				err = s.r.Flip.Create(ctx, entity.CreateFlip{
					To:      entity.FlipUp,
					Date:    ping.Date,
					CheckId: ping.CheckId,
				})
			}
		case entity.PingFail:
			err = s.r.Check.PingFail(ctx, ping.CheckId, ping.Date)
			if err != nil {
				return err
			}

			if status != entity.CheckDown {
				err = s.r.Flip.Create(ctx, entity.CreateFlip{
					To:      entity.FlipDown,
					Date:    ping.Date,
					CheckId: ping.CheckId,
				})
			}
		}
		if err != nil {
			return err
		}

		if ping.Type != entity.PingStart {
			lastPing, err := s.r.Ping.GetLastTypeAndDate(ctx, ping.CheckId)
			if err != nil {
				appErr, ok := err.(errors.AppError)
				if !ok || appErr.Kind != errors.NotExist {
					return err
				}
			}
			if lastPing != nil && lastPing.Type == entity.PingStart {
				ping.Duration.Int32 = int32(math.Round(ping.Date.Sub(lastPing.Date).Seconds()))
				ping.Duration.Valid = true
			}
		}

		err = s.r.Ping.Create(ctx, ping)
		if err != nil {
			return err
		}

		return nil
	})
}
