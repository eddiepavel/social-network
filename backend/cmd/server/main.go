package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"social-network/internal/routes"
	"social-network/pkg/db/sqlite"
)

func main() {
	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}

	// Configure database
	dbConfig := sqlite.Config{
		DatabasePath:   filepath.Join(workDir, "social-network.db"),
		MigrationsPath: filepath.Join(workDir, "pkg", "db", "migrations", "sqlite"),
	}

	// Initialize database with migrations
	db, err := sqlite.New(dbConfig)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	log.Println("Database migrations completed successfully")

	// Set up routes
	mux := routes.Setup(db.DB)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Wrap mux with CORS middleware
	handler := corsMiddleware(mux)

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// CORS middleware to allow frontend requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
