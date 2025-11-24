package routes

import (
	"net/http"
	"social-network/app"
	"social-network/internal/handlers"
)

type Handler struct {
	App     *app.App
	Handler *http.ServeMux
}

// Authentication endpoints
func (h *Handler) authRoutes() *http.ServeMux {

	h.Handler.HandleFunc("GET /session", handlers.GetSession(h.App))
	h.Handler.HandleFunc("POST /logout", handlers.Logout(h.App))

	return h.Handler
}

// public routes
func (h *Handler) publicRoutes() *http.ServeMux {

	h.Handler.HandleFunc("POST /register", handlers.Register(h.App))
	h.Handler.HandleFunc("POST /login", handlers.Login(h.App))

	return h.Handler
}
