package routes

import (
	"database/sql"
	"net/http"
	"strings"

	"social-network/internal/auth"
	"social-network/internal/handlers"
)

// Setup initializes all application routes
func Setup(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	usersHandler := handlers.NewUsersHandler(db)
	groupsHandler := handlers.NewGroupsHandler(db)
	followersHandler := handlers.NewFollowersHandler(db)

	// Initialize middleware
	authMiddleware := auth.AuthMiddleware(db)

	// Health check endpoint
	mux.HandleFunc("/health", healthCheck)

	// Authentication endpoints (public)
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)

	// Protected endpoints
	mux.HandleFunc("/api/auth/session", authMiddleware(authHandler.GetSession))

	// Users & Profile endpoints (protected)
	mux.HandleFunc("/api/users/", authMiddleware(usersHandler.GetUserProfile))
	mux.HandleFunc("/api/users/profile", authMiddleware(usersHandler.UpdateProfile))
	mux.HandleFunc("/api/users/privacy", authMiddleware(usersHandler.UpdatePrivacy))

	// Groups endpoints (protected)
	// Route to /api/groups for both GET (list all) and POST (create)
	mux.HandleFunc("/api/groups", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			groupsHandler.GetGroups(w, r)
		} else if r.Method == http.MethodPost {
			groupsHandler.CreateGroup(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Route to /api/groups/* for dynamic paths
	mux.HandleFunc("/api/groups/", authMiddleware(groupsRouter(groupsHandler)))

	// Followers endpoints (protected)
	mux.HandleFunc("/api/follow/requests", authMiddleware(followersHandler.GetFollowRequests))
	mux.HandleFunc("/api/follow/accept/", authMiddleware(followersHandler.AcceptFollowRequest))
	mux.HandleFunc("/api/follow/reject/", authMiddleware(followersHandler.RejectFollowRequest))
	mux.HandleFunc("/api/followers/", authMiddleware(followersHandler.GetFollowers))
	mux.HandleFunc("/api/following/", authMiddleware(followersHandler.GetFollowing))
	mux.HandleFunc("/api/follow/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			followersHandler.FollowUser(w, r)
		} else if r.Method == http.MethodDelete {
			followersHandler.UnfollowUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// TODO: Add more endpoints as features are implemented
	// Posts endpoints
	// mux.HandleFunc("/api/posts", authMiddleware(postsHandler.GetPosts))
	// mux.HandleFunc("/api/posts", authMiddleware(postsHandler.CreatePost))

	// Followers endpoints
	// mux.HandleFunc("/api/follow/:userId", authMiddleware(followersHandler.Follow))
	// mux.HandleFunc("/api/followers/:userId", authMiddleware(followersHandler.GetFollowers))

	// Chat WebSocket
	// mux.HandleFunc("/ws/chat", authMiddleware(chatHandler.HandleWebSocket))

	// Notifications endpoints
	// mux.HandleFunc("/api/notifications", authMiddleware(notificationsHandler.GetNotifications))

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
