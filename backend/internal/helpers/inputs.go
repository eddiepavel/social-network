package helpers

import (
	"database/sql"
	"errors"
	"net/http"
	"social-network/pkg/db/sqlite"
)

func ValidateRegister(r *http.Request, d *sql.DB) map[string][]interface{} {
	return map[string][]interface{}{
		"email": {"required", "string", func(v interface{}) error {
			email, _ := v.(string)
			_, err := sqlite.NewQuery(d).Users.GetUserByEmail(r.Context(), email)

			if !errors.Is(err, sql.ErrNoRows) {
				return errors.New("username exists")
			}

			return nil

		}},
		"password":   {"required", "string"},
		"first_name": {"required", "string"},
		"last_name":  {"required", "string"},
		"dob":        {"required", "string"},
		"avatar":     {"required", "string"},
		"nickname": {"required", "string", func(v interface{}) error {
			nickname, _ := v.(string)
			_, err := sqlite.NewQuery(d).Users.GetUserByNickname(r.Context(), nickname)

			if !errors.Is(err, sql.ErrNoRows) {
				return errors.New("nickname exists")
			}

			return nil

		}},
	}
}
