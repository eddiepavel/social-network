package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	contextkeys "social-network/internal/contextKeys"
	"social-network/internal/helpers"
	"social-network/internal/utils"
	db_sessions "social-network/pkg/db/queries/sessions"
	"social-network/pkg/db/sqlite"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// AuthMiddleware is middleware that requires a valid session
func (m *MiddlewareChain) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := GetSessionCookie(r)
		if err != nil {
			m.App.Logger.Error("cookie validation failed", "err", err)
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		userID, err := ValidateSession(m.App.DB, sessionID)
		if err != nil {
			m.App.Logger.Error("user validation failed", "err", err)
			utils.Unauthorized(w, "Unauthorized")
			//always delete session if expired after response
			if userID != nil {
				if err := InvalidateSession(m.App.DB, userID); err != nil {
					m.App.Logger.Error("failed to delete session", "err", err)
				}
			}
			return
		}

		// Attach user_id to request context
		ctx := context.WithValue(r.Context(), contextkeys.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserIDFromContext retrieves the user_id from the request context
func GetUserIDFromContext(ctx context.Context) ([]byte, bool) {
	userID, ok := ctx.Value(contextkeys.UserIDKey).([]byte)
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
		Name:     contextkeys.SessionCookieName,
		Value:    setUuid,
		Path:     "/",
		MaxAge:   int(contextkeys.SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: samesite, // Lax for localhost without HTTPS
		// Don't set Domain - let browser handle it automatically
	})
}

// GetSessionCookie retrieves the session cookie from the request
func GetSessionCookie(r *http.Request) ([]byte, error) {
	cookie, err := r.Cookie(contextkeys.SessionCookieName)
	if err != nil {
		return []byte{}, err
	}

	setUuid, err := helpers.GenerateFromString(cookie.Value)

	if err != nil {
		return nil, errors.New("invalid cookie")
	}

	return setUuid, nil
}

// ClearSessionCookie clears the session cookie (for logout)
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     contextkeys.SessionCookieName,
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
	expiresAt := time.Now().Add(contextkeys.SessionDuration)

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

	validate, err := sqlite.NewQuery(db).Sessions.ValidateSession(context.Background(), sessionID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}

	if !validate.Active.Bool || !validate.Active.Valid {
		return nil, fmt.Errorf("session is not active")
	}

	if time.Now().After(validate.ExpiresAt) && !errors.Is(err, sql.ErrNoRows) {
		return validate.SessionID, fmt.Errorf("session has expired")
	}

	return validate.UserID, nil
}

// InvalidateSession marks a session as inactive (for logout)
func InvalidateSession(db *sql.DB, sessionID []byte) error {
	err := sqlite.NewQuery(db).Sessions.InvalidateSession(context.Background(), sessionID)

	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}
