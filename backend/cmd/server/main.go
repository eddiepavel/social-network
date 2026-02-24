package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"social-network/app"
	"social-network/internal/routes"
	"social-network/internal/services"
	"social-network/internal/websocket"
	"social-network/pkg/db/sqlite"
	"social-network/pkg/environment"
)

func main() {
	//init dependencies need to start the app
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := environment.SetEnv(".env"); err != nil {
		log.Fatalf("Failed to start project %s: cp .env.example .env", err)
	}

	db, err := sqlite.New()

	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	logger.Info("database success")

	file := services.NewFileService("./storage/uploads", "./storage/public", db.DB, 5*time.Minute, logger)
	file.StartCleanUp()

	// Initialize WebSocket manager
	wsManager := websocket.NewManager(db.DB, logger, file.GenerateSignImage)
	wsManager.Start()
	defer wsManager.Shutdown()

	app := &app.App{
		DB:            db.DB,
		Logger:        logger,
		File:          file,
		WsManager:     wsManager,
		GuestSessions: make(map[string]time.Time),
	}

	defer func() {
		db.Close()
		file.StopCleanUp()
	}()

	// Set up routes
	mux := routes.Setup(app)

	port := os.Getenv("PORT")

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	serverError := make(chan error, 1)

	// Start server
	go func() {
		log.Printf("Server starting on port %s...", port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server failed to start:", err)
			serverError <- err
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-serverError:
		logger.Error("server error", "err", err)
	case sig := <-stop:
		logger.Info("receive signal", "info", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Info("server shutdown error", "err", err)
		return
	}

}
