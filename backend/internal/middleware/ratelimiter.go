package middleware

import (
	"errors"
	"net/http"
	contextkeys "social-network/internal/contextKeys"
	"social-network/internal/utils"
)

func (m *MiddlewareChain) RateLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cookie http.Cookie

		guest, err := r.Cookie(contextkeys.GuestSession)

		if err != nil {
			user, err := r.Cookie(contextkeys.SessionCookieName)
			if err != nil {
				utils.BadRequest(w, errors.New("where is your cookie cookieman?"))
				return
			}
			cookie = *user
		} else {
			cookie = *guest
		}

		allow := m.App.Rate.Allow(cookie.Value)

		if !allow && r.URL.Path != "/session" {
			utils.Error(w, 429, "429", "too many requests", "cooldown 120 seconds")
			return
		}

		next(w, r)
	}
}
