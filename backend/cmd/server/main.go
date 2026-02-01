package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"social-network/app"
	"social-network/internal/routes"
	"social-network/internal/services"
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

	file := services.NewFileService("./storage/uploads", db.DB, 5*time.Minute, logger)
	file.StartCleanUp()

	// Initialize WebSocket Hub
	hub := services.NewHub(logger)
	go hub.Run()

	app := &app.App{
		DB:     db.DB,
		Logger: logger,
		File:   file,
		Hub:    hub,
	}

	defer func() {
		db.Close()
		file.StopCleanUp()
	}()

	// Set up routes
	mux := routes.Setup(app)

	//start server
	port := os.Getenv("PORT")

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
