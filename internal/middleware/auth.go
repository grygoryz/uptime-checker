package middleware

import (
	"context"
	"gitlab.com/grygoryz/uptime-checker/internal/repository"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/respond"
	"net/http"
)

const SessionCookieName = "sessionId"

type userCtx struct{}

type UserSession struct {
	repository.UserSession
	SessionId string
}

// Auth middleware checks if user is authenticated and provides session data to request context
func Auth(session *repository.Session) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			sessionCookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				if err == http.ErrNoCookie {
					err = errors.E(errors.Unauthorized, "cookie not found", err)
				}
				respond.Error(r.Context(), w, err)
				return
			}

			user, err := session.Get(r.Context(), sessionCookie.Value)
			if err != nil {
				respond.Error(r.Context(), w, err)
				return
			}

			ctx := context.WithValue(r.Context(), userCtx{}, UserSession{
				SessionId:   sessionCookie.Value,
				UserSession: user,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func User(ctx context.Context) UserSession {
	user, ok := ctx.Value(userCtx{}).(UserSession)
	if !ok {
		return UserSession{}
	}

	return user
}
