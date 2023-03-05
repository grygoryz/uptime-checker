package repository

import (
	"context"
	"database/sql"
	goerrors "errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"time"
)

type checkRepository struct {
	db *sqlx.DB
}

func NewCheck(db *sqlx.DB) *checkRepository {
	return &checkRepository{db}
}

// GetMany returns user's checks
func (r *checkRepository) GetMany(ctx context.Context, userId int) ([]entity.Check, error) {
	q := getQueryable(ctx, r.db)
	var checks []entity.Check

	query := `SELECT id,
       "name",
       description,
       "interval",
       grace,
       last_ping,
       next_ping,
       last_started,
       status,
       (SELECT json_agg(json_build_object(
           'id', id,
           'kind', kind,
           'email', email,
           'webhook_url_up', webhook_url_up,
           'webhook_url_down', webhook_url_down
       )) channels
        FROM checks_channels
        INNER JOIN channels on checks_channels.channel_id = channels.id
        WHERE checks_channels.check_id = checks.id) channels
		FROM checks
		WHERE used_id = $1`
	err := q.SelectContext(ctx, &checks, query, userId)
	if err != nil {
		return nil, err
	}

	return checks, nil
}

// Get returns user's check by id
func (r *checkRepository) Get(ctx context.Context, params entity.GetCheck) (entity.Check, error) {
	q := getQueryable(ctx, r.db)
	var check entity.Check

	query := `SELECT id,
       "name",
       description,
       "interval",
       grace,
       last_ping,
       next_ping,
       last_started,
       status,
       (SELECT json_agg(json_build_object(
           'id', id,
           'kind', kind,
           'email', email,
           'webhook_url_up', webhook_url_up,
           'webhook_url_down', webhook_url_down
       ))
        FROM checks_channels
        INNER JOIN channels on checks_channels.channel_id = channels.id
        WHERE checks_channels.check_id = checks.id) channels
		FROM checks
		WHERE id = $1 AND used_id = $2`
	err := q.GetContext(ctx, &check, query, params.Id, params.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.E(errors.NotExist, "check not found")
		}
		return check, err
	}

	return check, nil
}

// Create creates check and returns its id
func (r *checkRepository) Create(ctx context.Context, check entity.CreateCheck) (string, error) {
	q := getQueryable(ctx, r.db)

	var id string
	query := `INSERT INTO checks ("name", description, "interval", grace, status, used_id)
	VALUES ($1, $2, $3, $4, 'new', $5)
	RETURNING id`
	err := q.
		QueryRowxContext(
			ctx,
			query,
			check.Name,
			check.Description,
			check.Interval,
			check.Grace,
			check.UserId,
		).
		Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

// Update updates check
func (r *checkRepository) Update(ctx context.Context, check entity.UpdateCheck) error {
	q := getQueryable(ctx, r.db)

	query := `UPDATE checks SET "name" = $1, description = $2, "interval" = $3, grace = $4 WHERE id = $5 AND used_id = $6`
	result, err := q.ExecContext(
		ctx, query, check.Name, check.Description, check.Interval, check.Grace, check.Id, check.UserId,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.E(errors.NotExist, "check not found")
	}

	return nil
}

// Delete deletes check
func (r *checkRepository) Delete(ctx context.Context, check entity.DeleteCheck) error {
	q := getQueryable(ctx, r.db)

	query := "DELETE FROM checks WHERE id = $1 AND used_id = $2"
	result, err := q.ExecContext(ctx, query, check.Id, check.UserId)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.E(errors.NotExist, "check not found")
	}

	return nil
}

// GetStatus returns check's status
func (r *checkRepository) GetStatus(ctx context.Context, checkId string) (entity.CheckStatus, error) {
	q := getQueryable(ctx, r.db)
	var status entity.CheckStatus

	query := `SELECT status FROM checks WHERE id = $1`
	err := q.GetContext(ctx, &status, query, checkId)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.E(errors.NotExist, "check not found")
		}
		return status, err
	}

	return status, nil
}

// SetStatus sets check's status
func (r *checkRepository) SetStatus(ctx context.Context, check entity.SetCheckStatus) error {
	q := getQueryable(ctx, r.db)

	query := `UPDATE checks SET "status" = $1 WHERE id = $2 AND used_id = $3`
	result, err := q.ExecContext(ctx, query, check.Status, check.Id, check.UserId)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.E(errors.NotExist, "check not found")
	}

	return nil
}

type checkChannel struct {
	CheckId   string `db:"check_id"`
	ChannelId int    `db:"channel_id"`
}

// AddChannels adds channels to check
func (r *checkRepository) AddChannels(ctx context.Context, params entity.AddChannels) error {
	q := getQueryable(ctx, r.db)

	rows := make([]checkChannel, len(params.Channels))
	for i, channelId := range params.Channels {
		rows[i] = checkChannel{CheckId: params.Id, ChannelId: channelId}
	}

	query := "INSERT INTO checks_channels (check_id, channel_id) VALUES (:check_id, :channel_id)"
	_, err := q.NamedExecContext(ctx, query, rows)
	if err != nil {
		var pgErr *pgconn.PgError
		if goerrors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return errors.E(errors.NotExist, "these channels does not exist")
		}

		return err
	}

	return nil
}

// DeleteChannels deletes check's channels
func (r *checkRepository) DeleteChannels(ctx context.Context, checkId string) error {
	q := getQueryable(ctx, r.db)

	result, err := q.ExecContext(ctx, "DELETE FROM checks_channels WHERE check_id = $1", checkId)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.E(errors.NotExist, "check not found")
	}

	return nil
}

// PingSuccess applies success ping to check
func (r *checkRepository) PingSuccess(ctx context.Context, checkId string, t time.Time) error {
	q := getQueryable(ctx, r.db)

	query := `UPDATE checks
	SET last_ping    = $1,
    	next_ping    = $1::timestamptz + (concat(interval, 's'))::interval,
    	last_started = NULL,
    	status       = 'up'
	WHERE id = $2`
	result, err := q.ExecContext(ctx, query, t, checkId)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.E(errors.NotExist, "check not found or it's paused")
	}

	return nil
}

// PingStart applies start ping to check
func (r *checkRepository) PingStart(ctx context.Context, checkId string, t time.Time) error {
	q := getQueryable(ctx, r.db)

	query := `UPDATE checks
	SET last_started = $1,
    	status       = 'started'
	WHERE id = $2`
	result, err := q.ExecContext(ctx, query, t, checkId)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.E(errors.NotExist, "check not found or it's paused")
	}

	return nil
}

// PingFail applies fail ping to check
func (r *checkRepository) PingFail(ctx context.Context, checkId string, t time.Time) error {
	q := getQueryable(ctx, r.db)

	query := `UPDATE checks
	SET last_ping    = $1,
	    next_ping    = NULL,
    	last_started = NULL,
    	status       = 'down'
	WHERE id = $2`
	result, err := q.ExecContext(ctx, query, t, checkId)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.E(errors.NotExist, "check not found or it's paused")
	}

	return nil
}

// GetExpired returns expired checks
func (r *checkRepository) GetExpired(ctx context.Context) ([]entity.CheckExpired, error) {
	q := getQueryable(ctx, r.db)
	var checks []entity.CheckExpired

	query := `SELECT
    ch.id,
    "name",
    next_ping, 
    grace,
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
	FROM checks ch
	INNER JOIN users u on u.id = ch.used_id
	WHERE status = 'up' AND current_timestamp > (next_ping + (concat(grace, 's'))::interval)
	FOR UPDATE SKIP LOCKED`
	err := q.SelectContext(ctx, &checks, query)
	if err != nil {
		return nil, err
	}

	return checks, nil
}

// SetDown sets checks status to down and next ping to status "down"
func (r *checkRepository) SetDown(ctx context.Context, checkIds []string) error {
	q := getQueryable(ctx, r.db)

	query, args, err := sqlx.In(`UPDATE checks
	SET next_ping    = NULL,
    	status       = 'down'
	WHERE id IN (?);`, checkIds)
	query = r.db.Rebind(query)
	_, err = q.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
