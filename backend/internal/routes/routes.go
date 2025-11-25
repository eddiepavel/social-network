package routes

import (
	"net/http"
	"social-network/internal/handlers"
)

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

func (h *Handler) userRoutes() *http.ServeMux {

	h.Handler.HandleFunc("GET /{id}", handlers.GetUserProfile(h.App))
	h.Handler.HandleFunc("PUT /profile", handlers.UpdateProfile(h.App))
	h.Handler.HandleFunc("PUT /privacy", handlers.UpdatePrivacy(h.App))

	return h.Handler
}
