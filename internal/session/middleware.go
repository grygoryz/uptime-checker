package session

import (
	"context"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/respond"
	"net/http"
)

const CookieName = "sessionId"

type userCtx struct{}

type userSession struct {
	UserData
	SessionId string
}

// Auth middleware checks if user is authenticated and provides session data to request context
func Auth(repository *Repository) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			sessionCookie, err := r.Cookie(CookieName)
			if err != nil {
				if err == http.ErrNoCookie {
					err = errors.E(errors.Unauthorized, "cookie not found", err)
				}
				respond.Error(r.Context(), w, err)
				return
			}

			user, err := repository.Get(r.Context(), sessionCookie.Value)
			if err != nil {
				respond.Error(r.Context(), w, err)
				return
			}

			ctx := context.WithValue(r.Context(), userCtx{}, userSession{
				SessionId: sessionCookie.Value,
				UserData:  user,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func User(ctx context.Context) userSession {
	user, ok := ctx.Value(userCtx{}).(userSession)
	if !ok {
		return userSession{}
	}

	return user
}
