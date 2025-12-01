package middleware

import (
	"net/http"
)

// LoggingMiddleware logs the details of each incoming HTTP request, such as method and path.
func (m *MiddlewareChain) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.App.Logger.Info("Request", "method", r.Method, "path", r.URL.Path)
		next(w, r)
	})
}
