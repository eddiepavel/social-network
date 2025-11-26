package middleware

import (
	"net/http"
)

// CORS middleware to allow frontend requests
func (m *MiddlewareChain) CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow specific origins
		allowedOrigins := map[string]bool{
			"http://localhost:5173":            true,
			"http://127.0.0.1:5173":            true,
			"https://fipwnwbppb.a.pinggy.link": true,
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	}
}
