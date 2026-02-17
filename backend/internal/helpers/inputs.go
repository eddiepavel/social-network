package helpers

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"social-network/app"
	contextkeys "social-network/internal/contextKeys"
	"social-network/pkg/db/sqlite"
	"time"
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
				return errors.New("email already exists")
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

			if action != "request" && action != "remove" && action != "accept_invite" && action != "decline_invite" {
				return errors.New("bad payload")
			}

			return nil
		}},
	}
}

type PostValidator struct{}

func (pv PostValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"content": {"required", "string", "min:1", "max:500"},
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
		"content": {"required", "string", "min:1", "max:500"},
		"image_id": {"sometimes", func(v interface{}) error {
			// Handle pointer type
			imageIdPtr, ok := v.(*string)
			if !ok || imageIdPtr == nil {
				// Field not provided, keep existing image
				return nil
			}

			uuId := *imageIdPtr

			// Empty string is allowed (to remove image)
			if uuId == "" {
				return nil
			}

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
		"image_id": {"sometimes", func(v interface{}) error {
			uuId := v.(string)

			// Empty string is allowed (no image)
			if uuId == "" {
				return nil
			}

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

type UpdateCommentValidator struct{}

func (ucv UpdateCommentValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"content": {"required", "string", "min:1", "max:500"},
		"image_id": {"sometimes", func(v interface{}) error {
			// Handle pointer type
			imageIdPtr, ok := v.(*string)
			if !ok || imageIdPtr == nil {
				// Field not provided, keep existing image
				return nil
			}

			uuId := *imageIdPtr

			// Empty string is allowed (to remove image)
			if uuId == "" {
				return nil
			}

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

type CreateGroupEventValidator struct{}

func (cgev CreateGroupEventValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"title":       {"required", "string"},
		"description": {"required", "string"},
		"timestamp": {"required", "string", func(v interface{}) error {
			date := v.(string)

			parsedTime, err := time.Parse(time.RFC3339, date)

			if err != nil {
				return errors.New("invalid date format, expected ISO 8601 (e.g., 2006-01-02T15:04:05Z)")
			}

			if parsedTime.Before(time.Now()) {
				return errors.New("event date must be in the future")
			}

			return nil
		}},
	}
}

type CreateNotificationValidator struct{}

func (cn CreateNotificationValidator) Build(r *http.Request, app *app.App) map[string][]interface{} {
	return map[string][]interface{}{
		"receiver_id": {"required", "string"},
		"type": {"required", "string", func(v interface{}) error {
			notifType := v.(string)
			validTypes := []string{"follow_request", "group_invitation", "group_request", "group_event", "message"}
			for _, t := range validTypes {
				if notifType == t {
					return nil
				}
			}
			return errors.New("invalid notification type")
		}},
		"from_id":  {"required", "string"},
		"group_id": {"sometimes", "string"},
		"event_id": {"sometimes", "string"},
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
	ValidateEventCreate          ValidationRuleBuilder = CreateGroupEventValidator{}
	ValidateCreateNotification   ValidationRuleBuilder = CreateNotificationValidator{}
)
