package routes

import (
	"net/http"
	"social-network/app"
	"social-network/internal/middleware"
)

type Handler struct {
	App *app.App
}

func (h *Handler) createGroup(routes func() *http.ServeMux, middlewares []string) http.HandlerFunc {

	builder := middleware.MiddlewareChain{
		App: h.App,
	}

	// Create the mux once, not on every request
	mux := routes()

	// Wrap ServeHTTP in a proper HandlerFunc
	handlerFunc := http.HandlerFunc(mux.ServeHTTP)

	group := builder.ChainMiddleware(
		handlerFunc,
		middlewares,
	)

	return group
}

func Setup(app *app.App) http.Handler {
	mux := http.NewServeMux()
	apiMux := http.NewServeMux()

	handler := &Handler{
		App: app,
	}

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	//group handlers
	authGroup := handler.createGroup(handler.authRoutes, []string{"auth"})
	publicGroup := handler.createGroup(handler.publicRoutes, []string{"guest"})
	usersGroup := handler.createGroup(handler.userRoutes, []string{"auth"})
	followersGroup := handler.createGroup(handler.followersRoutes, []string{"auth"})
	postsGroup := handler.createGroup(handler.postsRoutes, []string{"auth"})
	groupsGroup := handler.createGroup(handler.groupsRoutes, []string{"auth"})
	storageRoutes := handler.createGroup(handler.storageRoutes, []string{"auth"})
	chatRoutes := handler.createGroup(handler.chatRoutes, []string{"auth"})
	eventsRoutes := handler.createGroup(handler.eventsRoutes, []string{"auth"}) // No auth middleware, handled internally
	notificationsGroup := handler.createGroup(handler.notificationsRoutes, []string{"auth"})
	wsGroup := handler.createGroup(handler.wsRoutes, []string{"auth"})

	// prefix
	apiMux.Handle("/public/", http.StripPrefix("/public", publicGroup))
	apiMux.Handle("/auth/", http.StripPrefix("/auth", authGroup))
	apiMux.Handle("/users/", http.StripPrefix("/users", usersGroup))
	apiMux.Handle("/followers/", http.StripPrefix("/followers", followersGroup))
	apiMux.Handle("/posts/", http.StripPrefix("/posts", postsGroup))
	apiMux.Handle("/groups/", http.StripPrefix("/groups", groupsGroup))
	apiMux.Handle("/storage/", http.StripPrefix("/storage", storageRoutes))
	apiMux.Handle("/chat/", http.StripPrefix("/chat", chatRoutes))
	apiMux.Handle("/events/", http.StripPrefix("/events", eventsRoutes))
	apiMux.Handle("/notifications/", http.StripPrefix("/notifications", notificationsGroup))
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// WebSocket route (outside /api/ prefix)
	mux.Handle("/ws/", http.StripPrefix("/ws", wsGroup))

	return mux

}
