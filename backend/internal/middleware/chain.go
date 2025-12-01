package middleware

import (
	"fmt"
	"net/http"
	"social-network/app"
)

type Middleware func(h http.HandlerFunc) http.HandlerFunc

type MiddlewareChain struct {
	App *app.App
}

// ChainMiddleware applies a sequence of middlewares to an HTTP handler.
// It combines global middlewares with route-specific middlewares.
func (m *MiddlewareChain) ChainMiddleware(h http.HandlerFunc, k []string) http.HandlerFunc {

	selectMiddle := map[string]Middleware{
		"auth":    m.AuthMiddleware,
		"cors":    m.CorsMiddleware,
		"recover": m.RecoveryMiddleware,
		"log":     m.LoggingMiddleware,
	}

	globalMiddle := []string{"cors", "recover", "log"}

	wrapped := h

	fullMiddlewareList := append(globalMiddle, k...)

	for i := len(fullMiddlewareList) - 1; i >= 0; i-- {
		key := fullMiddlewareList[i]
		if mw, exists := selectMiddle[key]; exists {
			wrapped = mw(wrapped)
		} else {
			fmt.Printf("Middleware %s not found\n", key)
		}
	}

	return wrapped
}
