package session

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"time"
)

type Repository struct {
	redis *redis.Client
}

func NewRepository(redis *redis.Client) *Repository {
	return &Repository{redis}
}

type UserData struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

const TTL = time.Hour * 168

func (s *Repository) Create(ctx context.Context, data UserData) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	value, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	idStr := id.String()
	err = s.redis.Set(ctx, sessionKey(idStr), value, TTL).Err()
	if err != nil {
		return "", err
	}

	return idStr, nil
}

func (s *Repository) Get(ctx context.Context, id string) (UserData, error) {
	val, err := s.redis.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			err = errors.E(errors.NotExist, "session not found")
		}
		return UserData{}, err
	}

	var user UserData
	err = json.Unmarshal(val, &user)
	if err != nil {
		return UserData{}, err
	}

	return user, nil
}

func (s *Repository) Destroy(ctx context.Context, id string) error {
	return s.redis.Del(ctx, sessionKey(id)).Err()
}

func sessionKey(id string) string {
	return "sessions:" + id
}
