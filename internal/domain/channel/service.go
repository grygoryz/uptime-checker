package channel

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
)

type service struct {
	r *repository.Registry
}

func NewService(repositoryRegistry *repository.Registry) *service {
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
	return s.r.WithTx(ctx, func(ctx context.Context) error {
		ids, err := s.r.Channel.GetChecksDependentOnChannel(ctx, channel.Id)
		if err != nil {
			return err
		}
		if len(ids) > 0 {
			msg := fmt.Sprintf("there are checks that depend on this channel only: %v", ids)
			return errors.E(errors.Validation, msg)
		}

		err = s.r.Channel.Delete(ctx, channel)
		if err != nil {
			return err
		}

		return nil
	}, sql.LevelDefault)
}
