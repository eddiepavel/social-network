package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	db_followers "social-network/pkg/db/queries/followers"
	db_groups "social-network/pkg/db/queries/groups"
	db_image "social-network/pkg/db/queries/image"
	db_notifications "social-network/pkg/db/queries/notifications"
	db_posts "social-network/pkg/db/queries/posts"
	db_sessions "social-network/pkg/db/queries/sessions"
	db_users "social-network/pkg/db/queries/users"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection and provides methods for database operations
type DB struct {
	*sql.DB
}

// Config holds database configuration
type Config struct {
	DatabasePath   string
	MigrationsPath string
}

type Transactions struct {
	Users     *db_users.Queries
	Followers *db_followers.Queries
	Sessions  *db_sessions.Queries
	Posts     *db_posts.Queries
	Groups    *db_groups.Queries
	Image     *db_image.Queries
	Notifications *db_notifications.Queries
}

func NewQuery(db *sql.DB) *Transactions {
	return &Transactions{
		Users:     db_users.New(db),
		Followers: db_followers.New(db),
		Sessions:  db_sessions.New(db),
		Posts:     db_posts.New(db),
		Groups:    db_groups.New(db),
		Image:     db_image.New(db),
		Notifications: db_notifications.New(db),
	}
}

// New creates a new database connection and applies migrations
func New() (*DB, error) {
	// Ensure database directory exists
	dbDir := filepath.Dir("pkg/db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	fmt.Println("pkg/db/" + os.Getenv("DATABASE_NAME"))
	sqlDB, err := sql.Open("sqlite3", "pkg/db/"+os.Getenv("DATABASE_NAME"))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign key constraints
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{sqlDB}

	// Run init.sql first to create schema_migrations table
	if err := db.runInitSQL("pkg/db/migrations/sqlite"); err != nil {
		return nil, fmt.Errorf("failed to run init.sql: %w", err)
	}

	// Apply migrations
	if err := db.runMigrations("pkg/db/migrations/sqlite"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// runInitSQL executes the init.sql file to set up migration tracking
func (db *DB) runInitSQL(migrationsPath string) error {
	initPath := filepath.Join(migrationsPath, "init.sql")
	fmt.Println(initPath)
	// Read init.sql file
	initSQL, err := os.ReadFile(initPath)
	if err != nil {
		return fmt.Errorf("failed to read init.sql: %w", err)
	}

	// Execute init.sql
	if _, err := db.Exec(string(initSQL)); err != nil {
		return fmt.Errorf("failed to execute init.sql: %w", err)
	}

	log.Println("Executed init.sql successfully")
	return nil
}

// runMigrations applies all pending migrations from the up directory
func (db *DB) runMigrations(migrationsPath string) error {
	upPath := filepath.Join(migrationsPath, "up")

	// Read all migration files
	files, err := os.ReadDir(upPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files by name (they should be numbered)
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply each migration
	for _, filename := range migrationFiles {
		version := strings.TrimSuffix(filename, ".sql")

		// Check if migration already applied
		applied, err := db.isMigrationApplied(version)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", version, err)
		}

		if applied {
			log.Printf("Migration %s already applied, skipping", version)
			continue
		}

		// Read migration file
		migrationPath := filepath.Join(upPath, filename)
		migrationSQL, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Apply migration
		log.Printf("Applying migration: %s", version)
		if err := db.applyMigration(version, string(migrationSQL)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}

		log.Printf("Successfully applied migration: %s", version)
	}

	return nil
}

// isMigrationApplied checks if a migration has already been applied
func (db *DB) isMigrationApplied(version string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ? AND success = 1", version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// applyMigration executes a migration and records it in schema_migrations
func (db *DB) applyMigration(version, migrationSQL string) error {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migrationSQL); err != nil {
		// Record failed migration
		tx.Exec("INSERT INTO schema_migrations (version, success) VALUES (?, 0)", version)
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record successful migration
	if _, err := tx.Exec("INSERT INTO schema_migrations (version, success) VALUES (?, 1)", version); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

func CheckUniqueConstraint(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint")
}
