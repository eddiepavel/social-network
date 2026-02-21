package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	contextkeys "social-network/internal/contextKeys"
	"social-network/internal/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (m *MiddlewareChain) GuestSessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := m.GetGuestSession(r)

		if err != nil {
			utils.Error(w, 419, "419", "expired", "guest session expired")
			return
		}

		value, ok := m.App.GuestSessions[session]

		if !ok {
			utils.Error(w, 419, "419", "expired", "guest session expired")
			return
		}

		if time.Now().Add(1 * time.Minute).After(value) {
			delete(m.App.GuestSessions, session)

			utils.Error(w, 419, "419", "expired", "guest session expired")
			return
		}

		ctx := context.WithValue(r.Context(), contextkeys.GuestSession, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

}

func (m *MiddlewareChain) GetGuestSession(r *http.Request) (string, error) {

	session, err := r.Cookie(contextkeys.GuestSession)

	if err != nil {
		return "", errors.New("No cookie")
	}

	return session.Value, nil
}

func (m *MiddlewareChain) GenerateSessionID() (string, time.Duration) {
	uuId := uuid.New().String()

	sessionTime := time.Now().Add(5 * time.Minute)
	timeNow := time.Now()
	m.App.GuestSessions[uuId] = sessionTime

	return uuId, sessionTime.Sub(timeNow)
}

func (m *MiddlewareChain) CreateGuestSessionCookie(w http.ResponseWriter) string {

	uuId, time := m.GenerateSessionID()

	var secure bool
	var samesite http.SameSite

	isProduction, err := strconv.ParseBool(os.Getenv("PRODUCTION"))

	if err != nil {
		log.Fatal("failesd to read env variable")
	}

	if isProduction {
		secure = true
		samesite = http.SameSiteNoneMode
	} else {
		secure = false
		samesite = http.SameSiteLaxMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     contextkeys.GuestSession,
		Value:    uuId,
		Path:     "/",
		MaxAge:   int(time.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: samesite,
	})

	return uuId
}
