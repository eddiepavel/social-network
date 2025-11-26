package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"social-network/internal/helpers"
	"social-network/internal/utils"
	db_sessions "social-network/pkg/db/queries/sessions"
	"social-network/pkg/db/sqlite"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const SessionCookieName = "session_id"
const SessionDuration = 7 * 24 * time.Hour // 7 days

// AuthMiddleware is middleware that requires a valid session
func (m *MiddlewareChain) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := GetSessionCookie(r)
		if err != nil {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		userID, err := ValidateSession(m.App.DB, sessionID)
		if err != nil {
			m.App.Logger.Error("validation failed", "err", err)
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Attach user_id to request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserIDFromContext retrieves the user_id from the request context
func GetUserIDFromContext(ctx context.Context) ([]byte, bool) {
	userID, ok := ctx.Value(UserIDKey).([]byte)
	return userID, ok
}

// SetSessionCookie sets a session cookie in the response
func SetSessionCookie(w http.ResponseWriter, sessionID []byte) {
	setUuid, _ := helpers.GenerateFromBytes(sessionID)
	var secure bool
	var samesite http.SameSite
	production, err := strconv.ParseBool(os.Getenv("PRODUCTION"))
	if err != nil {
		log.Fatal("env corrupted")
	}

	if production {
		secure = true
		samesite = http.SameSiteNoneMode
	} else {
		secure = false
		samesite = http.SameSiteLaxMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    setUuid,
		Path:     "/",
		MaxAge:   int(SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: samesite, // Lax for localhost without HTTPS
		// Don't set Domain - let browser handle it automatically
	})
}

// GetSessionCookie retrieves the session cookie from the request
func GetSessionCookie(r *http.Request) ([]byte, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return []byte{}, err
	}
	setUuid, _ := helpers.GenerateFromString(cookie.Value)

	if err != nil {
		return []byte{}, err
	}
	return setUuid, nil
}

// ClearSessionCookie clears the session cookie (for logout)
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// CreateSession creates a new session for a user
func CreateSession(db *sql.DB, userID []byte) (db_sessions.Session, error) {
	sessionID, _ := uuid.New().MarshalBinary()
	expiresAt := time.Now().Add(SessionDuration)

	session, err := sqlite.NewQuery(db).Sessions.CreateSession(context.Background(),
		db_sessions.CreateSessionParams{
			SessionID: sessionID,
			UserID:    userID,
			ExpiresAt: expiresAt,
		})

	if err != nil {
		return db_sessions.Session{}, err
	}

	return session, nil
}

// ValidateSession checks if a session is valid and returns the user_id
func ValidateSession(db *sql.DB, sessionID []byte) ([]byte, error) {
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
func InvalidateSession(db *sql.DB, sessionID []byte) error {
	query := `DELETE FROM sessions WHERE session_id = ?`

	_, err := db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}
