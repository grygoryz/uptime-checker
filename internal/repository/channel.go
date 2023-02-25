package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
)

type channelRepository struct {
	db *sqlx.DB
}

func NewChannel(db *sqlx.DB) *channelRepository {
	return &channelRepository{db}
}

// Create creates channel of any kind
func (r *channelRepository) Create(ctx context.Context, channel entity.CreateChannel) (int, error) {
	q := getQueryable(ctx, r.db)

	var err error
	var id int
	switch channel.Kind {
	case entity.EmailChannel:
		query := "INSERT INTO channels (kind, email, user_id) VALUES ($1, $2, $3) RETURNING id"
		err = q.QueryRowxContext(ctx, query, channel.Kind, channel.Email, channel.UserId).Scan(&id)
	case entity.WebhookChannel:
		query := "INSERT INTO channels (kind, webhook_url, user_id) VALUES ($1, $2, $3) RETURNING id"
		err = q.QueryRowxContext(ctx, query, channel.Kind, channel.WebhookURL, channel.UserId).Scan(&id)
	default:
		return 0, fmt.Errorf("invalid channel kind: %v", channel.Kind)
	}

	if err != nil {
		return 0, err
	}
	return id, err
}

// Update updates channel by id
func (r *channelRepository) Update(ctx context.Context, channel entity.Channel) error {
	q := getQueryable(ctx, r.db)

	var result sql.Result
	var err error
	switch channel.Kind {
	case entity.EmailChannel:
		query := "UPDATE channels SET kind = $1, email = $2, webhook_url = null WHERE id = $3 AND user_id = $4"
		result, err = q.ExecContext(ctx, query, channel.Kind, channel.Email, channel.Id, channel.UserId)
	case entity.WebhookChannel:
		query := "UPDATE channels SET kind = $1, webhook_url = $2, email = null WHERE id = $3 AND user_id = $4"
		result, err = q.ExecContext(ctx, query, channel.Kind, channel.WebhookURL, channel.Id, channel.UserId)
	default:
		return fmt.Errorf("invalid channel kind: %v", channel.Kind)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.E(errors.NotExist, "channel not found")
	}

	return err
}

// GetMany returns user's channels
func (r *channelRepository) GetMany(ctx context.Context, userId int) ([]entity.ChannelShort, error) {
	q := getQueryable(ctx, r.db)
	var channels []entity.ChannelShort

	query := "SELECT id, kind, email, webhook_url FROM channels WHERE user_id = $1"
	err := q.SelectContext(ctx, &channels, query, userId)
	if err != nil {
		return nil, err
	}

	return channels, nil
}

// Delete deletes channel by id
func (r *channelRepository) Delete(ctx context.Context, channel entity.DeleteChannel) error {
	q := getQueryable(ctx, r.db)

	query := "DELETE FROM channels WHERE id = $1 AND user_id = $2"
	result, err := q.ExecContext(ctx, query, channel.Id, channel.UserId)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.E(errors.NotExist, "channel not found")
	}

	return nil
}

// CreateEmail creates email channel
func (r *channelRepository) CreateEmail(ctx context.Context, email string, userId int) error {
	q := getQueryable(ctx, r.db)
	query := "INSERT INTO channels (kind, email, user_id) VALUES ($1, $2, $3)"
	_, err := q.ExecContext(ctx, query, entity.EmailChannel, email, userId)
	if err != nil {
		return err
	}

	return nil
}

// GetChecksDependentOnChannel returns check ids that bound to this channel only
func (r *channelRepository) GetChecksDependentOnChannel(ctx context.Context, id int) ([]string, error) {
	q := getQueryable(ctx, r.db)

	query := `WITH checks_ids AS (SELECT check_id
	FROM checks_channels
	WHERE channel_id = $1)
	SELECT check_id
	FROM checks_channels
	WHERE check_id IN (SELECT check_id FROM checks_ids)
	GROUP BY check_id
	HAVING count(check_id) = 1`

	var ids []string
	err := q.SelectContext(ctx, &ids, query, id)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
