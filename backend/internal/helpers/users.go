package helpers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_followers "social-network/pkg/db/queries/followers"
	db_users "social-network/pkg/db/queries/users"
	"social-network/pkg/db/sqlite"
)

func FetchUser(app *app.App, userID []byte, ctx context.Context, w http.ResponseWriter) db_users.User {
	user, err := sqlite.NewQuery(app.DB).Users.GetUserById(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		utils.NotFound(w)
		return db_users.User{}
	}
	if err != nil {
		app.Logger.Error("Failed to fetch user", "error", err.Error())
		utils.Internal(w, err)
		return db_users.User{}
	}

	return user
}

func UpdateUser(user db_users.User, req models.UpdateProfileRequest) db_users.User {
	empty := 5
	if req.FirstName != nil && *req.FirstName != user.FirstName {
		empty--
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil && *req.LastName != user.LastName {
		empty--
		user.LastName = *req.LastName
	}
	if req.Avatar != nil && *req.Avatar != user.Avatar.String {
		empty--
		user.Avatar.String = *req.Avatar
		user.Avatar.Valid = true
	}
	if req.Nickname != nil && *req.Nickname != user.Nickname.String {
		empty--
		user.Nickname = sql.NullString{Valid: true, String: *req.Nickname}
	}
	if req.AboutMe != nil && *req.AboutMe != user.AboutMe.String {
		empty--
		user.AboutMe.String = *req.AboutMe
		user.AboutMe.Valid = true
	}

	if empty == 5 {
		return db_users.User{}
	}

	return user
}

// canViewProfile checks if currentUser can view targetUser's profile
// Rules:
// - Can always view own profile
// - Can view public profiles
// - Can view private profiles if following them (accepted follower)
func AccessToProfile(currentUserID []byte, targetUser db_users.User, app *app.App, r *http.Request) (bool, error) {
	// Can always view own profile
	if string(currentUserID) == string(targetUser.UserID) {
		return true, nil
	}

	// Public profiles are visible to everyone
	if targetUser.IsPublic {
		return true, nil
	}

	// For private profiles, check if current user is a follower
	_, err := sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(), db_followers.CheckIfUserFollowsParams{
		FollowerID: currentUserID, FolloweeID: targetUser.UserID,
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func UserToResponse(user db_users.User) models.UserResponse {
	setUuid, _ := GenerateFromBytes(user.UserID)
	return models.UserResponse{
		UserID:    setUuid,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		DOB:       user.Dob,
		Avatar: func() string {
			if user.Avatar.Valid {
				return user.Avatar.String
			}
			return ""
		}(),
		Nickname: func() string {
			if user.Nickname.Valid {
				return user.Nickname.String
			}
			return ""
		}(),
		AboutMe: func() string {
			if user.AboutMe.Valid {
				return user.AboutMe.String
			}
			return ""
		}(),
		IsPublic:  &user.IsPublic,
		CreatedAt: user.CreatedAt.Time.String(),
	}
}
