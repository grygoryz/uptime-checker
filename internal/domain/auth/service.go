package auth

import (
	"context"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/session"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	SignIn(ctx context.Context, user SignInBody) (string, error)
	SignUp(ctx context.Context, user SignUpBody) error
	SignOut(ctx context.Context, sessionId string) error
}

type service struct {
	r           *repository.Registry
	sessionRepo *session.Repository
}

func NewService(repositoryRegistry *repository.Registry, sessionRepo *session.Repository) Service {
	return &service{r: repositoryRegistry, sessionRepo: sessionRepo}
}

func (svc *service) SignUp(ctx context.Context, user SignUpBody) error {
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return svc.r.WithTx(ctx, func(ctx context.Context) error {
		id, err := svc.r.User.Create(ctx, user.Email, string(password))
		if err != nil {
			return err
		}

		err = svc.r.Channel.CreateEmail(ctx, user.Email, id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (svc *service) SignIn(ctx context.Context, user SignInBody) (string, error) {
	dbUser, err := svc.r.User.GetByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		return "", errors.E(errors.Unauthorized, "credentials are not valid")
	}

	id, err := svc.sessionRepo.Create(ctx, session.UserData{Id: dbUser.Id, Email: user.Email})
	if err != nil {
		return "", err
	}

	return id, nil
}

func (svc *service) SignOut(ctx context.Context, sessionId string) error {
	return svc.sessionRepo.Destroy(ctx, sessionId)
}
