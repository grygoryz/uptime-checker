package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"strings"
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

// GetUnprocessed returns unprocessed and not paused flips with corresponding check data
func (r *flipRepository) GetUnprocessed(ctx context.Context) ([]entity.FlipUnprocessed, error) {
	q := getQueryable(ctx, r.db)
	var flips []entity.FlipUnprocessed

	query := `SELECT 
    f.id,
    "date",
    "to",
    ch.name,
    u.email,
    (SELECT json_agg(json_build_object(
           'kind', kind,
           'email', email,
           'webhook_url_up', webhook_url_up,
           'webhook_url_down', webhook_url_down
    )) channels
    FROM checks_channels
    INNER JOIN channels on checks_channels.channel_id = channels.id
    WHERE checks_channels.check_id = ch.id) channels
    FROM flips f
    INNER JOIN checks ch on ch.id = f.check_id
    INNER JOIN users u on u.id = ch.used_id
	WHERE processed = false AND "to" IN ('up', 'down')
	ORDER BY date
	FOR UPDATE SKIP LOCKED`
	err := q.SelectContext(ctx, &flips, query)
	if err != nil {
		return nil, err
	}

	return flips, nil
}

// CreateMany creates many flips and returns their ids
func (r *flipRepository) CreateMany(ctx context.Context, flips []entity.CreateFlip) ([]int, error) {
	q := getQueryable(ctx, r.db)

	var qb strings.Builder
	qb.WriteString(`INSERT INTO flips ("to", "date", check_id) VALUES `)
	params := make([]interface{}, 0, len(flips)*3)
	for _, flip := range flips {
		qb.WriteString(`(?, ?, ?),`)
		params = append(params, flip.To, flip.Date, flip.CheckId)
	}
	query := qb.String()
	// rebind and remove trailing comma
	query = q.Rebind(query[:len(query)-1] + "RETURNING id")

	rows, err := q.QueryxContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}

	var ids []int
	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		ids = append(ids, id)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *flipRepository) SetProcessed(ctx context.Context, flipIds []int) error {
	q := getQueryable(ctx, r.db)

	query, args, err := sqlx.In(`UPDATE flips
	SET processed = true
	WHERE id IN (?)`, flipIds)
	query = r.db.Rebind(query)
	_, err = q.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
