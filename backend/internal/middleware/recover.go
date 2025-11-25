package middleware

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"social-network/internal/utils"
)

func RecoveryMiddleware(next http.HandlerFunc, db *sql.DB, l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				l.Error("server panic", "method", r.Method, "paht", r.URL.Path, "stacktrace", err)
				utils.Internal(w, errors.New("internal server error"))
			}
		}()

		next(w, r)
	}
}
