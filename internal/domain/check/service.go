package check

import (
	"context"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"time"
)

type Service interface {
	GetChecks(ctx context.Context, userId int) ([]entity.Check, error)
	GetCheck(ctx context.Context, userId int, checkId string) (entity.Check, error)
	CreateCheck(ctx context.Context, check entity.CreateCheck, channels []int) (string, error)
	UpdateCheck(ctx context.Context, check entity.UpdateCheck, channels []int) error
	DeleteCheck(ctx context.Context, check entity.DeleteCheck) error
	PauseCheck(ctx context.Context, checkId string, userId int) error
	GetPings(ctx context.Context, params entity.GetPings) ([]entity.Ping, int, error)
	GetFlips(ctx context.Context, params entity.GetFlips) ([]entity.Flip, int, error)
}

type service struct {
	r *repository.Registry
}

func NewService(repositoryRegistry *repository.Registry) Service {
	return &service{r: repositoryRegistry}
}

func (s *service) GetChecks(ctx context.Context, userId int) ([]entity.Check, error) {
	return s.r.Check.GetMany(ctx, userId)
}

func (s *service) GetCheck(ctx context.Context, userId int, checkId string) (entity.Check, error) {
	return s.r.Check.Get(ctx, entity.GetCheck{Id: checkId, UserId: userId})
}

func (s *service) CreateCheck(ctx context.Context, check entity.CreateCheck, channels []int) (string, error) {
	var id string
	err := s.r.WithTx(ctx, func(ctx context.Context) error {
		var err error
		id, err = s.r.Check.Create(ctx, check)
		if err != nil {
			return err
		}

		err = s.r.Check.AddChannels(ctx, entity.AddChannels{Id: id, Channels: channels})
		if err != nil {
			return err
		}

		return nil
	})

	return id, err
}

func (s *service) UpdateCheck(ctx context.Context, check entity.UpdateCheck, channels []int) error {
	return s.r.WithTx(ctx, func(ctx context.Context) error {
		err := s.r.Check.Update(ctx, check)
		if err != nil {
			return err
		}

		err = s.r.Check.DeleteChannels(ctx, check.Id)
		if err != nil {
			return err
		}

		err = s.r.Check.AddChannels(ctx, entity.AddChannels{Id: check.Id, Channels: channels})
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *service) DeleteCheck(ctx context.Context, check entity.DeleteCheck) error {
	return s.r.Check.Delete(ctx, check)
}

func (s *service) PauseCheck(ctx context.Context, checkId string, userId int) error {
	return s.r.WithTx(ctx, func(ctx context.Context) error {
		status, err := s.r.Check.GetStatus(ctx, checkId)
		if status == entity.CheckPaused {
			return errors.E(errors.Validation, "check is paused already")
		}

		err = s.r.Check.SetStatus(ctx, entity.SetCheckStatus{
			Id:     checkId,
			UserId: userId,
			Status: entity.CheckPaused,
		})
		if err != nil {
			return err
		}

		err = s.r.Flip.Create(ctx, entity.CreateFlip{
			To:      entity.FlipPaused,
			Date:    time.Now(),
			CheckId: checkId,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *service) GetPings(ctx context.Context, params entity.GetPings) ([]entity.Ping, int, error) {
	var pings []entity.Ping
	var total int
	err := s.r.WithTx(ctx, func(ctx context.Context) error {
		var err error
		total, err = s.r.Ping.GetTotal(ctx, entity.GetPingsTotal{
			CheckId: params.CheckId,
			From:    params.From,
			To:      params.To,
		})
		if err != nil {
			return err
		}

		pings, err = s.r.Ping.GetMany(ctx, params)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return pings, total, nil
}

func (s *service) GetFlips(ctx context.Context, params entity.GetFlips) ([]entity.Flip, int, error) {
	var flips []entity.Flip
	var total int
	err := s.r.WithTx(ctx, func(ctx context.Context) error {
		var err error
		total, err = s.r.Flip.GetTotal(ctx, entity.GetFlipsTotal{
			CheckId: params.CheckId,
			From:    params.From,
			To:      params.To,
		})
		if err != nil {
			return err
		}

		flips, err = s.r.Flip.GetMany(ctx, params)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return flips, total, nil
}
