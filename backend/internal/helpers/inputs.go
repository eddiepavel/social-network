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
		"email": {"required", "email", func(v interface{}) error {
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

type LoginValidator struct{}

func (l LoginValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"email":    {"required", "email"},
		"password": {"required", "string"},
	}
}

type PrivacyValidator struct{}

func (p PrivacyValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"is_public": {"required", "boolean"},
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

type CreateGroupValidator struct{}

func (c CreateGroupValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"group_name": {"required", "string", func(v interface{}) error {
			groupName := v.(string)
			_, err := sqlite.NewQuery(app.DB).Groups.GetGroupByName(r.Context(), groupName)

			if !errors.Is(err, sql.ErrNoRows) {
				return errors.New("group name already exists")
			}
			return nil
		}},
		"description": {"required", "string", "min:10", "max:50"},
		"image":       {"sometimes"},
	}
}

type MemberShipGroupValidator struct{}

func (up MemberShipGroupValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"action": {"required", "string", func(v interface{}) error {
			action := v.(string)

			if action != "request" && action != "remove" {
				return errors.New("bad payload")
			}

			return nil
		}},
	}
}

type PostValidator struct{}

func (pv PostValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"content": {"required", "string", "min:10", "max:255"},
		"visibility": {"required", "string", func(v interface{}) error {
			visibility := v.(string)
			if visibility != "public" && visibility != "private" && visibility != "semi-private" {
				return errors.New("invalid post visibility")
			}
			return nil
		}},
		"image_id": {"sometimes", "string"},
	}
}

type UpdatePostValidator struct{}

func (upv UpdatePostValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"content":  {"required", "string", "min:10", "max:255"},
		"image_id": {"sometimes", "string"},
	}
}

type UpdatePostVisibilityValidator struct{}

func (upvv UpdatePostVisibilityValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"visibility": {"required", "string", func(v interface{}) error {
			visibility := v.(string)
			if visibility != "public" && visibility != "private" && visibility != "semi-private" {
				return errors.New("invalid post visibility")
			}
			return nil
		}},
	}
}

type AddUserToPrivatePostValidator struct{}

func (aupv AddUserToPrivatePostValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"user_id": {"required", "string"},
	}
}

// Exported instances
var (
	ValidateRegister             ValidationRuleBuilder = RegisterValidator{}
	ValidateUpdateProfile        ValidationRuleBuilder = UpdateProfileValidator{}
	ValidateLogin                ValidationRuleBuilder = LoginValidator{}
	ValidatePrivacy              ValidationRuleBuilder = PrivacyValidator{}
	ValidateCreateGroup          ValidationRuleBuilder = CreateGroupValidator{}
	ValidateMemberShip           ValidationRuleBuilder = MemberShipGroupValidator{}
	ValidatePost                 ValidationRuleBuilder = PostValidator{}
	ValidateUpdatePost           ValidationRuleBuilder = UpdatePostValidator{}
	ValidateUpdatePostVisibility ValidationRuleBuilder = UpdatePostVisibilityValidator{}
	ValidateAddUserToPrivatePost ValidationRuleBuilder = AddUserToPrivatePostValidator{}
)
