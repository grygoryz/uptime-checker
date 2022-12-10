package middleware

import (
	"github.com/go-chi/cors"
	"net/http"
)

func CORS(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{})(next)
}
