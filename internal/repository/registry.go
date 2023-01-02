package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type Registry struct {
	db      *sqlx.DB
	User    User
	Channel Channel
}

func NewRegistry(db *sqlx.DB) *Registry {
	return &Registry{
		db:      db,
		User:    NewUser(db),
		Channel: NewChannel(db),
	}
}

type txFn func(ctx context.Context) (interface{}, error)

type txKey struct{}

// WithTx wraps function in transaction by providing *sqlx.Tx into the context. Transactions work only for
// repositories of the Registry
func (r *Registry) WithTx(ctx context.Context, fn txFn) (interface{}, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	res, err := fn(txCtx)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			return nil, err
		}

		return res, err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return nil, errCommit
	}

	return res, err
}

// getQueryable returns *sql.Tx if it exists in the context, otherwise returns *sqlx.DB
func getQueryable(ctx context.Context, db *sqlx.DB) queryable {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return db
}
