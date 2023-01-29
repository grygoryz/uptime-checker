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
	GetTotal(ctx context.Context, params entity.GetPingsTotal) (int, error)
	GetMany(ctx context.Context, params entity.GetPings) ([]entity.Ping, error)
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

// GetTotal returns check's pings total number for specified period
func (r *pingRepository) GetTotal(ctx context.Context, params entity.GetPingsTotal) (int, error) {
	q := getQueryable(ctx, r.db)
	var total int

	query := `SELECT count(*)
    FROM pings
	WHERE check_id = $1 AND date >= $2 AND date <= $3`
	err := q.GetContext(ctx, &total, query, params.CheckId, params.From, params.To)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// GetMany returns check's pings
func (r *pingRepository) GetMany(ctx context.Context, params entity.GetPings) ([]entity.Ping, error) {
	q := getQueryable(ctx, r.db)
	var pings []entity.Ping

	query := `SELECT id, "type", "date", source, user_agent, duration, body
    FROM pings
	WHERE check_id = $1 AND date >= $2 AND date <= $3
	ORDER BY date DESC
	LIMIT $4 OFFSET $5`
	err := q.SelectContext(ctx, &pings, query, params.CheckId, params.From, params.To, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	return pings, nil
}
