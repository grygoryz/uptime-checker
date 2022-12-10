package auth

import (
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
)

type Service interface {
	SignIn() (string, error)
}

type service struct {
	userRepository repository.User
}

func NewService(userRepository repository.User) Service {
	return &service{userRepository}
}

func (svc *service) SignIn() (string, error) {
	return "Auth success", nil
}
