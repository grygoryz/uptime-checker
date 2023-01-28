package channel

import (
	"context"
	"fmt"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
)

type Service interface {
	CreateChannel(ctx context.Context, channel entity.CreateChannel) (int, error)
	UpdateChannel(ctx context.Context, channel entity.Channel) error
	GetChannels(ctx context.Context, userId int) ([]entity.ChannelShort, error)
	DeleteChannel(ctx context.Context, channel entity.DeleteChannel) error
}

type service struct {
	r *repository.Registry
}

func NewService(repositoryRegistry *repository.Registry) Service {
	return &service{r: repositoryRegistry}
}

func (s *service) CreateChannel(ctx context.Context, channel entity.CreateChannel) (int, error) {
	return s.r.Channel.Create(ctx, channel)
}

func (s *service) UpdateChannel(ctx context.Context, channel entity.Channel) error {
	return s.r.Channel.Update(ctx, channel)
}

func (s *service) GetChannels(ctx context.Context, userId int) ([]entity.ChannelShort, error) {
	return s.r.Channel.GetMany(ctx, userId)
}

func (s *service) DeleteChannel(ctx context.Context, channel entity.DeleteChannel) error {
	_, err := s.r.WithTx(ctx, func(ctx context.Context) (interface{}, error) {
		ids, err := s.r.Channel.GetChecksDependentOnChannel(ctx, channel.Id)
		if err != nil {
			return nil, err
		}
		if len(ids) > 0 {
			msg := fmt.Sprintf("there are checks that depend on this channel only: %v", ids)
			return nil, errors.E(errors.Validation, msg)
		}

		err = s.r.Channel.Delete(ctx, channel)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}
