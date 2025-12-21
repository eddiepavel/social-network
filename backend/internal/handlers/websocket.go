package handlers

import (
	"net/http"
	"social-network/app"
	"social-network/internal/middleware"
	"social-network/internal/utils"
)

// ConnectWebSocket handles the WebSocket upgrade request
func ConnectWebSocket(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by auth middleware)
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Check if WebSocket manager is available
		if app.WsManager == nil {
			app.Logger.Error("WebSocket manager not initialized")
			http.Error(w, "WebSocket service unavailable", http.StatusServiceUnavailable)
			return
		}

		// Delegate to WebSocket manager for upgrade and client handling
		app.WsManager.ServeWs(w, r, userID)
	}
}
