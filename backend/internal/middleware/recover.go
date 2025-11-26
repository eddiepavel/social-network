package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"
	"social-network/internal/utils"
)

func (m *MiddlewareChain) RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				m.App.Logger.Error("server panic", "method", r.Method, "path", r.URL.Path, "stacktrace_message", err, "stacktrace", string(stack))
				utils.Internal(w, errors.New("internal server error"))
			}
		}()

		next(w, r)
	}
}
