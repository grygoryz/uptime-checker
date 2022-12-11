package respond

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog"
	"gitlab.com/grygoryz/uptime-checker/internal/middleware"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"net/http"
)

func Error(ctx context.Context, w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	log := middleware.LogEntry(ctx)

	switch e := err.(type) {
	case errors.AppError:
		if e.Err != nil {
			log.Error().Err(e.Err).Msg("underlying error")
		}

		var status int
		switch e.Kind {
		case errors.Forbidden:
			status = http.StatusForbidden
		case errors.NotExist:
			status = http.StatusNotFound
		case errors.Validation:
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}

		w.WriteHeader(status)
		write(w, e.Error(), log)
	default:
		log.Error().Err(e).Msg("unhandled error")

		w.WriteHeader(http.StatusInternalServerError)
		write(w, "Internal error", log)
	}

}

type response struct {
	Message string `json:"message"`
}

func write(w http.ResponseWriter, msg string, log zerolog.Logger) {
	m := response{msg}
	body, err := json.Marshal(m)
	if err != nil {
		log.Error().Err(err).Send()
	}

	_, err = w.Write(body)
	if err != nil {
		log.Error().Err(err).Send()
	}
}
