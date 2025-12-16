package routes

import (
	"net/http"
	"social-network/internal/handlers"
	"social-network/internal/middleware"
)

// Authentication endpoints
func (h *Handler) authRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /session", handlers.GetSession(h.App))
	mux.HandleFunc("POST /logout", handlers.Logout(h.App))

	return mux
}

// public routes
func (h *Handler) publicRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", handlers.Register(h.App))
	mux.HandleFunc("POST /login", handlers.Login(h.App))

	return mux
}

func (h *Handler) userRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /profile/{id}", handlers.GetUserProfile(h.App))
	mux.HandleFunc("PUT /profile", handlers.UpdateProfile(h.App))
	mux.HandleFunc("PUT /privacy", handlers.UpdatePrivacy(h.App))
	mux.HandleFunc("GET /search", handlers.SearchUsers(h.App))

	return mux
}

func (h *Handler) followersRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /requests", handlers.GetFollowRequests(h.App))
	mux.HandleFunc("POST /requests/{requestId}/respond", handlers.UpdateFollowRequest(h.App))
	mux.HandleFunc("POST /{userId}/follow", handlers.FollowUser(h.App))
	mux.HandleFunc("GET /{userId}/followers", handlers.GetFollowers(h.App))
	mux.HandleFunc("GET /{userId}/following", handlers.GetFollowing(h.App))

	return mux
}

func (h *Handler) postsRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /create", handlers.CreatePost(h.App))
	mux.HandleFunc("GET /feed", handlers.GetFeedPosts(h.App))

	mux.HandleFunc("GET /id/{postId}", handlers.GetPostWithCommentsReactions(h.App))
	mux.HandleFunc("PUT /id/{postId}", handlers.EditPost(h.App))
	mux.HandleFunc("DELETE /id/{postId}", handlers.DeletePost(h.App))

	mux.HandleFunc("PUT /id/{postId}/privacy", handlers.UpdatePostVisibility(h.App))
	mux.HandleFunc("POST /id/{postId}/privacy", handlers.AddUserToPrivatePostList(h.App))
	mux.HandleFunc("DELETE /id/{postId}/privacy", handlers.RemoveUserFromPrivatePostList(h.App))

	mux.HandleFunc("GET /id/{postId}/comment", handlers.GetComments(h.App))
	mux.HandleFunc("POST /id/{postId}/comment", handlers.CreateComment(h.App))
	mux.HandleFunc("PUT /id/{postId}/comment/{commentId}", handlers.EditComment(h.App))
	mux.HandleFunc("DELETE /id/{postId}/comment/{commentId}", handlers.DeleteComment(h.App))
	mux.HandleFunc("POST /id/{postId}/comment/{commentId}/reaction", handlers.ToggleReaction(h.App))

	mux.HandleFunc("GET /id/{postId}/reaction", handlers.GetReactions(h.App))
	mux.HandleFunc("POST /id/{postId}/reaction", handlers.ToggleReaction(h.App))

	return mux
}

func (h *Handler) groupsRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /group/{groupId}", handlers.GetGroup(h.App))
	mux.HandleFunc("GET /all", handlers.GetGroups(h.App))
	mux.HandleFunc("POST /create", handlers.CreateGroup(h.App))
	mux.HandleFunc("DELETE /delete/{groupId}", handlers.DeleteGroup(h.App))
	mux.HandleFunc("PUT /update/{groupId}", handlers.UpdateGroup(h.App))
	mux.HandleFunc("POST /invite/{groupId}", handlers.InviteToGroup(h.App))
	mux.HandleFunc("POST /members/request/{groupId}", handlers.RequestToJoinGroup(h.App))
	mux.HandleFunc("GET /members/requests/{groupId}", handlers.GetGroupRequests(h.App)) //group admin route only
	mux.HandleFunc("POST /members/respond/{groupId}", handlers.RespondRequest(h.App))   //group admin route only
	mux.HandleFunc("POST /members/remove/{groupId}", handlers.RemoveMember(h.App))      //group admin route or if current logged in user want to leave joined group

	return mux
}

func (h *Handler) storageRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /upload", handlers.Upload(h.App))
	mux.HandleFunc("GET /image/{image}", handlers.GetImage(h.App))

	return mux
}

func (h *Handler) notificationsRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Read-only endpoints for frontend
	mux.HandleFunc("GET /", handlers.GetNotifications(h.App))
	mux.HandleFunc("GET /unseen", handlers.GetUnseenNotifications(h.App))
	mux.HandleFunc("GET /details", handlers.GetNotificationsWithUserDetails(h.App))
	mux.HandleFunc("GET /unseen/count", handlers.GetUnseenNotificationCount(h.App))
	mux.HandleFunc("PUT /{notificationId}/seen", handlers.MarkNotificationAsSeen(h.App))
	mux.HandleFunc("PUT /seen/all", handlers.MarkAllNotificationsAsSeen(h.App))
	mux.HandleFunc("DELETE /{notificationId}", handlers.DeleteNotification(h.App))

	return mux
}

func (h *Handler) wsRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Test endpoint to verify auth middleware is working
	mux.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "No user in context", 401)
			return
		}
		w.Write([]byte("Auth middleware works! User ID: " + string(userID)))
	})

	// WebSocket connection
	mux.Handle("GET /connect", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if h.App.WsManager != nil {
			h.App.WsManager.ServeWs(w, r, userID)
		} else {
			http.Error(w, "WebSocket service unavailable", http.StatusServiceUnavailable)
		}
	}))

	return mux
}
