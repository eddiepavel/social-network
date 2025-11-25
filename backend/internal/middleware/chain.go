package middleware

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
)

type Middleware func(h http.HandlerFunc, c *sql.DB, l *slog.Logger) http.HandlerFunc

// ChainMiddleware applies a sequence of middlewares to an HTTP handler.
// It combines global middlewares with route-specific middlewares.
func ChainMiddleware(h http.HandlerFunc, k []string, c *sql.DB, l *slog.Logger) http.HandlerFunc {

	selectMiddle := map[string]Middleware{
		"auth":    AuthMiddleware,
		"cors":    CorsMiddleware,
		"recover": RecoveryMiddleware,
	}

	globalMiddle := []string{"cors", "recover"}

	wrapped := h

	fullMiddlewareList := append(globalMiddle, k...)

	for i := len(fullMiddlewareList) - 1; i >= 0; i-- {
		key := fullMiddlewareList[i]
		if mw, exists := selectMiddle[key]; exists {
			wrapped = mw(wrapped, c, l)
		} else {
			fmt.Printf("Middleware %s not found\n", key)
		}
	}

	return wrapped
}
