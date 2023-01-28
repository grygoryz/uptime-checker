package repository

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
)

type Ping interface {
	Create(ctx context.Context, ping entity.CreatePing) error
	GetLastTypeAndDate(ctx context.Context, checkId string) (*entity.PingTypeAndDate, error)
}

type pingRepository struct {
	db *sqlx.DB
}

func NewPing(db *sqlx.DB) *pingRepository {
	return &pingRepository{db}
}

// Create creates ping
func (r *pingRepository) Create(ctx context.Context, ping entity.CreatePing) error {
	q := getQueryable(ctx, r.db)

	query := `INSERT INTO pings 
   ("type", source, user_agent, duration, body, check_id, "date")
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := q.ExecContext(
		ctx,
		query,
		ping.Type,
		ping.Source,
		ping.UserAgent,
		ping.Duration,
		ping.Body,
		ping.CheckId,
		ping.Date,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetLastTypeAndDate returns last ping's type and date
func (r *pingRepository) GetLastTypeAndDate(ctx context.Context, checkId string) (*entity.PingTypeAndDate, error) {
	q := getQueryable(ctx, r.db)
	var ping entity.PingTypeAndDate

	query := `SELECT "type", "date"
    FROM pings
	WHERE check_id = $1
	ORDER BY date DESC
	LIMIT 1`
	err := q.GetContext(ctx, &ping, query, checkId)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.E(errors.NotExist, "ping not found")
		}
		return nil, err
	}

	return &ping, nil
}
