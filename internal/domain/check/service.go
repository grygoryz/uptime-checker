package check

import (
	"context"
	"fmt"
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
	id, err := s.r.WithTx(ctx, func(ctx context.Context) (interface{}, error) {
		id, err := s.r.Check.Create(ctx, check)
		if err != nil {
			return nil, err
		}

		err = s.r.Check.AddChannels(ctx, entity.AddChannels{Id: id, Channels: channels})
		if err != nil {
			return nil, err
		}

		return id, nil
	})

	return fmt.Sprintf("%v", id), err
}

func (s *service) UpdateCheck(ctx context.Context, check entity.UpdateCheck, channels []int) error {
	_, err := s.r.WithTx(ctx, func(ctx context.Context) (interface{}, error) {
		err := s.r.Check.Update(ctx, check)
		if err != nil {
			return nil, err
		}

		err = s.r.Check.DeleteChannels(ctx, check.Id)
		if err != nil {
			return nil, err
		}

		err = s.r.Check.AddChannels(ctx, entity.AddChannels{Id: check.Id, Channels: channels})
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (s *service) DeleteCheck(ctx context.Context, check entity.DeleteCheck) error {
	return s.r.Check.Delete(ctx, check)
}

func (s *service) PauseCheck(ctx context.Context, checkId string, userId int) error {
	_, err := s.r.WithTx(ctx, func(ctx context.Context) (interface{}, error) {
		status, err := s.r.Check.GetStatus(ctx, checkId)
		if status == entity.CheckPaused {
			return nil, errors.E(errors.Validation, "check is paused already")
		}

		err = s.r.Check.SetStatus(ctx, entity.SetCheckStatus{
			Id:     checkId,
			UserId: userId,
			Status: entity.CheckPaused,
		})
		if err != nil {
			return nil, err
		}

		err = s.r.Flip.Create(ctx, entity.CreateFlip{
			To:      entity.FlipPaused,
			Date:    time.Now(),
			CheckId: checkId,
		})
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

type pingsTxResult struct {
	Pings []entity.Ping
	Total int
}

func (s *service) GetPings(ctx context.Context, params entity.GetPings) ([]entity.Ping, int, error) {
	result, err := s.r.WithTx(ctx, func(ctx context.Context) (interface{}, error) {
		total, err := s.r.Ping.GetTotal(ctx, entity.GetPingsTotal{
			CheckId: params.CheckId,
			From:    params.From,
			To:      params.To,
		})
		if err != nil {
			return nil, err
		}

		pings, err := s.r.Ping.GetMany(ctx, params)
		if err != nil {
			return nil, err
		}

		return pingsTxResult{Pings: pings, Total: total}, nil
	})
	if err != nil {
		return nil, 0, err
	}

	txResult := result.(pingsTxResult)
	return txResult.Pings, txResult.Total, nil
}

type flipsTxResult struct {
	Flips []entity.Flip
	Total int
}

func (s *service) GetFlips(ctx context.Context, params entity.GetFlips) ([]entity.Flip, int, error) {
	result, err := s.r.WithTx(ctx, func(ctx context.Context) (interface{}, error) {
		total, err := s.r.Flip.GetTotal(ctx, entity.GetFlipsTotal{
			CheckId: params.CheckId,
			From:    params.From,
			To:      params.To,
		})
		if err != nil {
			return nil, err
		}

		flips, err := s.r.Flip.GetMany(ctx, params)
		if err != nil {
			return nil, err
		}

		return flipsTxResult{Flips: flips, Total: total}, nil
	})
	if err != nil {
		return nil, 0, err
	}

	txResult := result.(flipsTxResult)
	return txResult.Flips, txResult.Total, nil
}
