package repository

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"time"
)

type Session struct {
	redis *redis.Client
}

func NewSession(redis *redis.Client) *Session {
	return &Session{redis}
}

type UserSession struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

const SessionTTL = time.Hour * 168

func (s *Session) Create(ctx context.Context, data UserSession) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	value, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	idStr := id.String()
	err = s.redis.Set(ctx, sessionKey(idStr), value, SessionTTL).Err()
	if err != nil {
		return "", err
	}

	return idStr, nil
}

func (s *Session) Get(ctx context.Context, id string) (UserSession, error) {
	val, err := s.redis.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			err = errors.E(errors.NotExist, "session not found")
		}
		return UserSession{}, err
	}

	var user UserSession
	err = json.Unmarshal(val, &user)
	if err != nil {
		return UserSession{}, err
	}

	return user, nil
}

func (s *Session) Destroy(ctx context.Context, id string) error {
	return s.redis.Del(ctx, sessionKey(id)).Err()
}

func sessionKey(id string) string {
	return "sessions:" + id
}
