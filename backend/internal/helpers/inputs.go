package helpers

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"social-network/app"
	contextkeys "social-network/internal/contextKeys"
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
			existingUser, err := sqlite.NewQuery(app.DB).Users.GetUserByNickname(r.Context(), sql.NullString{Valid: true, String: nickname})

			if errors.Is(err, sql.ErrNoRows) {
				// Nickname is available
				return nil
			}
			if err != nil {
				return errors.New("something went wrong")
			}

			// Check if the nickname belongs to the current user (that's OK)
			currentUserID := r.Context().Value(contextkeys.UserIDKey).([]byte)
			if bytes.Equal(existingUser.UserID, currentUserID) {
				return nil
			}

			return errors.New("nickname exists")
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
		"image": {"sometimes", "string", func(v interface{}) error {
			value := v.(string)

			image, err := sqlite.NewQuery(app.DB).Image.GetImageById(r.Context(), value)

			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errors.New("wrong id")
				}

				return errors.New("something went wrong")
			}

			user := r.Context().Value(contextkeys.UserIDKey).([]byte)

			if !bytes.Equal(user, image.PosterID) {
				return errors.New("wrong id")
			}

			if !image.ExpiresAt.Valid {
				return errors.New("wrong id")
			}

			return nil
		}},
	}
}

type UpdateGroupValidator struct{}

func (c UpdateGroupValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"group_name": {"required", "string", func(v interface{}) error {
			groupName := v.(string)

			group, err := sqlite.NewQuery(app.DB).Groups.GetGroupByName(r.Context(), groupName)

			groupId, _ := GenerateFromString(r.PathValue("groupId"))

			if !errors.Is(err, sql.ErrNoRows) && !bytes.Equal(group.GroupID, groupId) {
				return errors.New("group name already exists")
			}
			return nil
		}},
		"description": {"required", "string", "min:10", "max:50"},
		"image": {"sometimes", "string", func(v interface{}) error {
			value := v.(string)

			image, err := sqlite.NewQuery(app.DB).Image.GetImageById(r.Context(), value)

			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errors.New("wrong id")
				}

				return errors.New("something went wrong")
			}

			user := r.Context().Value(contextkeys.UserIDKey).([]byte)

			if !bytes.Equal(user, image.PosterID) {
				return errors.New("wrong id")
			}

			if !image.ExpiresAt.Valid {
				return errors.New("wrong id")
			}

			return nil
		}},
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
		"image_id": {"sometimes", "string", func(v interface{}) error {
			uuId := v.(string)

			getImage, err := sqlite.NewQuery(app.DB).Image.GetImageById(r.Context(), uuId)

			user := r.Context().Value(contextkeys.UserIDKey).([]byte)

			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errors.New("image not found")
				}
				return errors.New("something went wrong")
			}

			if !bytes.Equal(getImage.PosterID, user) {
				return errors.New("image does not belong to you")
			}

			if getImage.ExpiresAt.Valid == false {
				return errors.New("image already assigned")
			}

			return nil
		}},
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

type CreateCommentValidator struct{}

func (ccv CreateCommentValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"content":   {"required", "string", "min:1", "max:500"},
		"parent_id": {"sometimes", "string"},
		// TODO: Add image validation when image service is ready
		"image_id": {"sometimes", "string"},
	}
}

type UpdateCommentValidator struct{}

func (ucv UpdateCommentValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"content": {"required", "string", "min:1", "max:500"},
	}
}

type MessageValidator struct{}

func (mv MessageValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"content": {"required", "string", "min:1", "max:500"},
	}
}

type FirstMessageValidator struct{}

func (fmv FirstMessageValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"content": {"required", "string", "min:1", "max:500"},
		"target_id": {"required", "string", func(v interface{}) error {
			targetId := v.(string)
			userID, err := GenerateFromString(targetId)
			if err != nil {
				return errors.New("wrong target id")
			}
			_, err = sqlite.NewQuery(app.DB).Users.GetUserById(r.Context(), userID)
			if err != nil {
				return errors.New("user not found")
			}
			return nil
		}},
	}
}

// Exported instances
var (
	ValidateRegister             ValidationRuleBuilder = RegisterValidator{}
	ValidateUpdateProfile        ValidationRuleBuilder = UpdateProfileValidator{}
	ValidateLogin                ValidationRuleBuilder = LoginValidator{}
	ValidatePrivacy              ValidationRuleBuilder = PrivacyValidator{}
	ValidateCreateGroup          ValidationRuleBuilder = CreateGroupValidator{}
	ValidateUpdateGroup          ValidationRuleBuilder = UpdateGroupValidator{}
	ValidateMemberShip           ValidationRuleBuilder = MemberShipGroupValidator{}
	ValidatePost                 ValidationRuleBuilder = PostValidator{}
	ValidateUpdatePost           ValidationRuleBuilder = UpdatePostValidator{}
	ValidateUpdatePostVisibility ValidationRuleBuilder = UpdatePostVisibilityValidator{}
	ValidateAddUserToPrivatePost ValidationRuleBuilder = AddUserToPrivatePostValidator{}
	ValidateCreateComment        ValidationRuleBuilder = CreateCommentValidator{}
	ValidateUpdateComment        ValidationRuleBuilder = UpdateCommentValidator{}
	ValidateMessage              ValidationRuleBuilder = MessageValidator{}
	ValidateFirstMessage         ValidationRuleBuilder = FirstMessageValidator{}
)
