package auth

import (
	"context"
	"database/sql"
	"log"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware is middleware that requires a valid session
func AuthMiddleware(db *sql.DB) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sessionID, err := GetSessionCookie(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := ValidateSession(db, sessionID)
			if err != nil {
				log.Printf("Session validation failed: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Attach user_id to request context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// GetUserIDFromContext retrieves the user_id from the request context
func GetUserIDFromContext(ctx context.Context) ([]byte, bool) {
	userID, ok := ctx.Value(UserIDKey).([]byte)
	return userID, ok
}
