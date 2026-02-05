package middleware

import (
	"errors"
	"net/http"
	"os"
	"social-network/internal/utils"
	"strconv"
)

// CORS middleware to allow frontend requests
func (m *MiddlewareChain) CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		allowedOrigin := os.Getenv("ALLOW_ORIGIN")

		isProduction, err := strconv.ParseBool(os.Getenv("PRODUCTION"))

		if err != nil {
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if !isProduction {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if isProduction && allowedOrigin == origin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next(w, r)
	}
}
