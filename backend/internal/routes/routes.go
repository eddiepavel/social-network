package routes

import (
	"net/http"
	"social-network/app"
	"social-network/internal/handlers"
	"social-network/internal/middleware"
	"strings"
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

// Setup initializes all application routes
func Setup(app *app.App) http.Handler {
	mux := http.NewServeMux()

	handler := &Handler{
		App:     app,
		Handler: mux,
	}

	authRoutesMux := handler.authRoutes()
	authWithMiddleware := middleware.ChainMiddleware(
		authRoutesMux.ServeHTTP,
		[]string{"auth"},
		app.DB,
		app.Logger,
	)

	publicRoutesMux := handler.publicRoutes()
	publicWithMiddleware := middleware.ChainMiddleware(
		publicRoutesMux.ServeHTTP,
		[]string{},
		app.DB,
		app.Logger,
	)

	mux.Handle("/api/", http.StripPrefix("/api", mux))
	mux.Handle("/public/", http.StripPrefix("/public", http.HandlerFunc(publicWithMiddleware)))
	mux.Handle("/auth/", http.StripPrefix("/auth", http.HandlerFunc(authWithMiddleware)))

	// Progress so far:
	// 1. Grouped endpoints with middleware for easier route management.
	// 2. Planning to move query logic from handlers/middleware to pkg/db using sqlc (see new YAML file).
	// 3. Added .env helper for consistent environment variable usage.
	// 4. Database migration logic needs refactoring.... migration paths should be fixed and not depend on execution locations.
	// 5. App struct that wraps entire application with packages that are needed throughout request lifecycle.

	// Disclaimer:
	// The app requires significant refactoring for better organization (aiming for a mini-framework structure).
	// Error responses should be standardized to JSON for all REST API endpoints, instead of plain text.

	// Initialize handlers
	usersHandler := handlers.NewUsersHandler(app.DB)
	groupsHandler := handlers.NewGroupsHandler(app.DB)
	followersHandler := handlers.NewFollowersHandler(app.DB)

	// Health check endpoint
	mux.HandleFunc("/health", healthCheck)

	// Users & Profile endpoints (protected)
	mux.HandleFunc("/api/users/", middleware.AuthMiddleware(usersHandler.GetUserProfile, app.DB, app.Logger))
	mux.HandleFunc("/api/users/profile", middleware.AuthMiddleware(usersHandler.UpdateProfile, app.DB, app.Logger))
	mux.HandleFunc("/api/users/privacy", middleware.AuthMiddleware(usersHandler.UpdatePrivacy, app.DB, app.Logger))

	// Groups endpoints (protected)
	// Route to /api/groups for both GET (list all) and POST (create)
	mux.HandleFunc("/api/groups", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			groupsHandler.GetGroups(w, r)
		} else if r.Method == http.MethodPost {
			groupsHandler.CreateGroup(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}, app.DB, app.Logger))

	// Route to /api/groups/* for dynamic paths
	mux.HandleFunc("/api/groups/", middleware.AuthMiddleware(groupsRouter(groupsHandler), app.DB, app.Logger))

	// Followers endpoints (protected)
	mux.HandleFunc("/api/follow/requests", middleware.AuthMiddleware(followersHandler.GetFollowRequests, app.DB, app.Logger))
	mux.HandleFunc("/api/follow/accept/", middleware.AuthMiddleware(followersHandler.AcceptFollowRequest, app.DB, app.Logger))
	mux.HandleFunc("/api/follow/reject/", middleware.AuthMiddleware(followersHandler.RejectFollowRequest, app.DB, app.Logger))
	mux.HandleFunc("/api/followers/", middleware.AuthMiddleware(followersHandler.GetFollowers, app.DB, app.Logger))
	mux.HandleFunc("/api/following/", middleware.AuthMiddleware(followersHandler.GetFollowing, app.DB, app.Logger))
	mux.HandleFunc("/api/follow/", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			followersHandler.FollowUser(w, r)
		} else if r.Method == http.MethodDelete {
			followersHandler.UnfollowUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}, app.DB, app.Logger))

	// TODO: Add more endpoints as features are implemented

	return mux
}

// healthCheck is a simple health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// groupsRouter routes group-related requests to the appropriate handler
// Handles paths like:
// GET  /api/groups/:id
// POST /api/groups/:id/invite
// POST /api/groups/:id/request
// POST /api/groups/:id/accept/:userId
func groupsRouter(h *handlers.GroupsHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse path: /api/groups/{groupId}/{action}/{userId}
		path := strings.TrimPrefix(r.URL.Path, "/api/groups/")
		parts := strings.Split(path, "/")

		// Filter out empty parts
		var filteredParts []string
		for _, part := range parts {
			if part != "" {
				filteredParts = append(filteredParts, part)
			}
		}

		if len(filteredParts) == 0 {
			http.Error(w, "Group ID required", http.StatusBadRequest)
			return
		}

		// Route based on path structure
		switch len(filteredParts) {
		case 1:
			// /api/groups/:id - GET group details
			if r.Method == http.MethodGet {
				h.GetGroupDetails(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}

		case 2:
			// /api/groups/:id/invite - POST invite user
			// /api/groups/:id/request - POST request to join
			action := filteredParts[1]
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			switch action {
			case "invite":
				h.InviteUser(w, r)
			case "request":
				h.RequestToJoin(w, r)
			default:
				http.Error(w, "Invalid action", http.StatusBadRequest)
			}

		case 3:
			// /api/groups/:id/accept/:userId - POST accept/reject join request
			action := filteredParts[1]
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			if action == "accept" {
				h.HandleJoinRequest(w, r)
			} else {
				http.Error(w, "Invalid action", http.StatusBadRequest)
			}

		default:
			http.Error(w, "Invalid path", http.StatusBadRequest)
		}
	}
}
