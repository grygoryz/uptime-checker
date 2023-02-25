package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
)

type flipRepository struct {
	db *sqlx.DB
}

func NewFlip(db *sqlx.DB) *flipRepository {
	return &flipRepository{db}
}

// Create creates flip
func (r *flipRepository) Create(ctx context.Context, flip entity.CreateFlip) error {
	q := getQueryable(ctx, r.db)

	query := `INSERT INTO flips ("to", "date", check_id) VALUES ($1, $2, $3)`
	_, err := q.ExecContext(ctx, query, flip.To, flip.Date, flip.CheckId)
	if err != nil {
		return err
	}

	return nil
}

// GetTotal returns check's flips total number for specified period
func (r *flipRepository) GetTotal(ctx context.Context, params entity.GetFlipsTotal) (int, error) {
	q := getQueryable(ctx, r.db)
	var total int

	query := `SELECT count(*)
    FROM flips
	WHERE check_id = $1 AND date >= $2 AND date <= $3`
	err := q.GetContext(ctx, &total, query, params.CheckId, params.From, params.To)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// GetMany returns check's flips
func (r *flipRepository) GetMany(ctx context.Context, params entity.GetFlips) ([]entity.Flip, error) {
	q := getQueryable(ctx, r.db)
	var flips []entity.Flip

	query := `SELECT "date", "to"
    FROM flips
	WHERE check_id = $1 AND date >= $2 AND date <= $3
	ORDER BY date DESC
	LIMIT $4 OFFSET $5`
	err := q.SelectContext(ctx, &flips, query, params.CheckId, params.From, params.To, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	return flips, nil
}
