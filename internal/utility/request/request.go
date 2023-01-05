package request

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/errors"
	"gitlab.com/grygoryz/uptime-checker/internal/validate"
	"net/http"
	"strconv"
)

func Body[T interface{}](r *http.Request, validator *validate.Validator) (T, error) {
	var result T
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		return result, err
	}

	err = validator.Struct(result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func IntParam(r *http.Request, param string) (int, error) {
	value, err := strconv.Atoi(chi.URLParam(r, param))
	if err != nil {
		return 0, errors.E(errors.Validation, fmt.Sprintf("param %q must be an int", param))
	}

	return value, err
}
