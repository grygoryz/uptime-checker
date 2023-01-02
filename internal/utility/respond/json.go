package respond

import (
	"context"
	"encoding/json"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/logger"
	"net/http"
)

func JSON(ctx context.Context, w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if payload == nil {
		w.WriteHeader(statusCode)
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		Error(ctx, w, err)
		return
	}

	w.WriteHeader(statusCode)

	if string(data) == "null" {
		return
	}

	_, err = w.Write(data)
	if err != nil {
		log := logger.LogEntry(ctx)
		log.Error().Err(err).Send()
	}
}
