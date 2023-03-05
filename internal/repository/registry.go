package repository

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type Registry struct {
	db      *sqlx.DB
	User    *userRepository
	Channel *channelRepository
	Check   *checkRepository
	Ping    *pingRepository
	Flip    *flipRepository
}

func NewRegistry(db *sqlx.DB) *Registry {
	return &Registry{
		db:      db,
		User:    NewUser(db),
		Channel: NewChannel(db),
		Check:   NewCheck(db),
		Ping:    NewPing(db),
		Flip:    NewFlip(db),
	}
}

type txFn func(ctx context.Context) error

type txKey struct{}

// WithTx wraps function in transaction by providing *sqlx.Tx into the context. Transactions work only for
// repositories of the Registry
func (r *Registry) WithTx(ctx context.Context, fn txFn, level sql.IsolationLevel) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	err = fn(txCtx)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			return err
		}

		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	return err
}

// getQueryable returns *sql.Tx if it exists in the context, otherwise returns *sqlx.DB
func getQueryable(ctx context.Context, db *sqlx.DB) queryable {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return db
}
