package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
)

type Channel interface {
	CreateEmail(ctx context.Context, email string, userId int) error
}

type channelRepository struct {
	db *sqlx.DB
}

func NewChannel(db *sqlx.DB) *channelRepository {
	return &channelRepository{db}
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
