package helpers

import (
	"database/sql"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/pkg/db/sqlite"
)

// ValidationRuleBuilder interface ensures all validators follow the same contract
type ValidationRuleBuilder interface {
	Build(r *http.Request, app *app.App) map[string][]interface{}
}

// RegisterValidator implements ValidationRuleBuilder for registration
type RegisterValidator struct{}

func (rv RegisterValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"email": {"required", "email", "string", func(v interface{}) error {
			email, _ := v.(string)
			_, err := sqlite.NewQuery(app.DB).Users.GetUserByEmail(r.Context(), email)

			if !errors.Is(err, sql.ErrNoRows) {
				return errors.New("username exists")
			}
			return nil
		}},
		"password":   {"required", "string"},
		"first_name": {"required", "string"},
		"last_name":  {"required", "string"},
		"dob":        {"required", "string"},
		"avatar":     {"sometimes", "base64"},
		"nickname": {"sometimes", "string", func(v interface{}) error {
			nickname, _ := v.(string)
			_, err := sqlite.NewQuery(app.DB).Users.GetUserByNickname(r.Context(), sql.NullString{Valid: true, String: nickname})

			if !errors.Is(err, sql.ErrNoRows) {
				return errors.New("nickname exists")
			}
			return nil
		}},
	}
}

// UpdateProfileValidator implements ValidationRuleBuilder for profile updates
type UpdateProfileValidator struct{}

func (upv UpdateProfileValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"first_name": {"sometimes", "string"},
		"last_name":  {"sometimes", "string"},
		"nickname": {"sometimes", "string", func(v interface{}) error {
			nickname, _ := v.(string)
			_, err := sqlite.NewQuery(app.DB).Users.GetUserByNickname(r.Context(), sql.NullString{Valid: true, String: nickname})

			if !errors.Is(err, sql.ErrNoRows) {
				return errors.New("nickname exists")
			}
			return nil
		}},
		"avatar":   {"sometimes", "base64"},
		"about_me": {"sometimes", "string"},
	}
}

// Exported instances
var (
	ValidateRegister      ValidationRuleBuilder = RegisterValidator{}
	ValidateUpdateProfile ValidationRuleBuilder = UpdateProfileValidator{}
)
