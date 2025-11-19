package auth

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const SessionDuration = 7 * 24 * time.Hour // 7 days

// CreateSession creates a new session for a user
func CreateSession(db *sql.DB, userID []byte) (string, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(SessionDuration)

	query := `
		INSERT INTO sessions (session_id, user_id, active, expires_at)
		VALUES (?, ?, 1, ?)
	`

	_, err := db.Exec(query, sessionID, userID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

// ValidateSession checks if a session is valid and returns the user_id
func ValidateSession(db *sql.DB, sessionID string) ([]byte, error) {
	var userID []byte
	var active bool
	var expiresAt time.Time

	query := `
		SELECT user_id, active, expires_at
		FROM sessions
		WHERE session_id = ?
	`

	err := db.QueryRow(query, sessionID).Scan(&userID, &active, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !active {
		return nil, fmt.Errorf("session is not active")
	}

	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("session has expired")
	}

	return userID, nil
}

// InvalidateSession marks a session as inactive (for logout)
func InvalidateSession(db *sql.DB, sessionID string) error {
	query := `UPDATE sessions SET active = 0 WHERE session_id = ?`

	_, err := db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}
