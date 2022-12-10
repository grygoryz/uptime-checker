package repository

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/errors"
)

type User interface {
	Get(ctx context.Context, id int) (*entity.User, error)
	Create(ctx context.Context, email string, password string) (int, error)
}

type repository struct {
	db *sqlx.DB
}

func NewUser(db *sqlx.DB) *repository {
	return &repository{db}
}

func (r *repository) Get(ctx context.Context, id int) (*entity.User, error) {
	user := entity.User{}
	err := r.db.GetContext(ctx, &user, "SELECT id, email FROM users WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.E(errors.NotExist, "User not found")
		}
		return nil, err
	}

	return &user, nil
}

// Create creates user and returns his id
func (r *repository) Create(ctx context.Context, email string, password string) (int, error) {
	var id int
	err := r.db.
		QueryRowxContext(ctx, "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id", email, password).
		Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}
