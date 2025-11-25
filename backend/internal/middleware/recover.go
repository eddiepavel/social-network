package middleware

import (
	"errors"
	"net/http"
	"social-network/internal/utils"
)

func (m *MiddlewareChain) RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.App.Logger.Error("server panic", "method", r.Method, "path", r.URL.Path, "stacktrace", err)
				utils.Internal(w, errors.New("internal server error"))
			}
		}()

		next(w, r)
	}
}
