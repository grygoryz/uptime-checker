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
)

type userRepository struct {
	db *sqlx.DB
}

func NewUser(db *sqlx.DB) *userRepository {
	return &userRepository{db}
}

// Create creates user and returns his id
func (r *userRepository) Create(ctx context.Context, email string, password string) (int, error) {
	q := getQueryable(ctx, r.db)
	var id int
	err := q.
		QueryRowxContext(ctx, "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id", email, password).
		Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if goerrors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, errors.E(errors.Duplicated, "user with this email exists already")
		}

		return 0, err
	}

	return id, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	q := getQueryable(ctx, r.db)
	user := entity.User{}
	err := q.GetContext(ctx, &user, "SELECT id, password FROM users WHERE email = $1", email)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.E(errors.NotExist, "user not found")
		}
		return entity.User{}, err
	}

	return user, nil
}
