package auth

import (
	"context"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	SignIn(ctx context.Context, user SignInBody) (string, error)
	SignUp(ctx context.Context, user SignUpBody) error
	SignOut(ctx context.Context, sessionId string) error
}

type service struct {
	r       *repository.Registry
	session *repository.Session
}

func NewService(repositoryRegistry *repository.Registry, session *repository.Session) Service {
	return &service{r: repositoryRegistry, session: session}
}

func (svc *service) SignUp(ctx context.Context, user SignUpBody) error {
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = svc.r.WithTx(ctx, func(ctx context.Context) (interface{}, error) {
		id, err := svc.r.User.Create(ctx, user.Email, string(password))
		if err != nil {
			return nil, err
		}

		err = svc.r.Channel.CreateEmail(ctx, user.Email, id)
		if err != nil {
			return nil, err
		}

		return nil, err
	})

	return err
}

func (svc *service) SignIn(ctx context.Context, user SignInBody) (string, error) {
	dbUser, err := svc.r.User.GetByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		return "", errors.E(errors.Unauthorized, "Credentials are not valid")
	}

	id, err := svc.session.Create(ctx, repository.UserSession{Id: dbUser.Id, Email: user.Email})
	if err != nil {
		return "", err
	}

	return id, nil
}

func (svc *service) SignOut(ctx context.Context, sessionId string) error {
	return svc.session.Destroy(ctx, sessionId)
}
