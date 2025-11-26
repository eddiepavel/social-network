package routes

import (
	"net/http"
	"social-network/app"
	"social-network/internal/handlers"
	"social-network/internal/middleware"
	"strings"
)

type Handler struct {
	App *app.App
}

func (h *Handler) createGroup(routes func() *http.ServeMux, middlewares []string) http.HandlerFunc {

	builder := middleware.MiddlewareChain{
		App: h.App,
	}

	group := builder.ChainMiddleware(
		routes().ServeHTTP,
		middlewares,
	)

	return group
}

func Setup(app *app.App) http.Handler {
	mux := http.NewServeMux()

	handler := &Handler{
		App: app,
	}

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	authGroup := handler.createGroup(handler.authRoutes, []string{"auth"})
	publicGroup := handler.createGroup(handler.publicRoutes, []string{})
	usersGroup := handler.createGroup(handler.userRoutes, []string{"auth"})
	followersGroup := handler.createGroup(handler.followersRoutes, []string{"auth"})

	mux.Handle("/api/", http.StripPrefix("/api", mux))
	mux.Handle("/public/", http.StripPrefix("/public", publicGroup))
	mux.Handle("/auth/", http.StripPrefix("/auth", authGroup))
	mux.Handle("/users/", http.StripPrefix("/users", usersGroup))
	mux.Handle("/followers/", http.StripPrefix("/followers", followersGroup))

	// Initialize handlers
	groupsHandler := handlers.NewGroupsHandler(app.DB)

	// Groups endpoints (protected)
	// Route to /api/groups for both GET (list all) and POST (create)
	bb := middleware.MiddlewareChain{
		App: app,
	}
	mux.HandleFunc("/api/groups", bb.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			groupsHandler.GetGroups(w, r)
		} else if r.Method == http.MethodPost {
			groupsHandler.CreateGroup(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}, []string{"auth"}))

	// Route to /api/groups/* for dynamic paths

	mux.HandleFunc("/api/groups/", bb.ChainMiddleware(groupsRouter(groupsHandler), []string{"auth"}))

	// TODO: Add more endpoints as features are implemented

	return mux
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
