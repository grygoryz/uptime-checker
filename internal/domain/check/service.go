package check

import (
	"context"
	"fmt"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
)

type Service interface {
	GetChecks(ctx context.Context, userId int) ([]entity.Check, error)
	GetCheck(ctx context.Context, userId int, checkId string) (entity.Check, error)
	CreateCheck(ctx context.Context, check entity.CreateCheck, channels []int) (string, error)
	UpdateCheck(ctx context.Context, check entity.UpdateCheck, channels []int) error
	DeleteCheck(ctx context.Context, check entity.DeleteCheck) error
	PauseCheck(ctx context.Context, checkId string, userId int) error
	ResumeCheck(ctx context.Context, checkId string, userId int) error
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
	return s.r.Check.SetStatus(ctx, entity.SetCheckStatus{
		Id:     checkId,
		UserId: userId,
		Status: entity.CheckPaused,
	})
}

func (s *service) ResumeCheck(ctx context.Context, checkId string, userId int) error {
	return s.r.Check.SetStatus(ctx, entity.SetCheckStatus{
		Id:     checkId,
		UserId: userId,
		Status: entity.CheckNew,
	})
}
