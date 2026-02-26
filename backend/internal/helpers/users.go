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
	"time"
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
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     &user.Email,
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

func UserToResponseImage(user db_users.User, app *app.App, userId []byte) models.UserResponse {
	setUuid, _ := GenerateFromBytes(user.UserID)
	email := user.Email
	return models.UserResponse{
		UserID:    setUuid,
		Email:     &email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		DOB:       user.Dob,
		Avatar: func() string {
			if user.Avatar.Valid && user.Avatar.String != "" {
				return app.File.GenerateSignImage(user.Avatar.String, userId, time.Now().Add(15*time.Minute))
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
		AvatarID:  user.Avatar.String,
	}
}

func UserToResponseProfile(user db_users.GetUserByIdWithCountsRow, app *app.App, userId []byte, permission bool, isOwnProfile bool) models.UserResponse {
	setUuid, _ := GenerateFromBytes(user.UserID)
	var userResp models.UserResponse
	userResp = models.UserResponse{
		UserID:    setUuid,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email: func() *string {
			if isOwnProfile {
				return &user.Email
			}
			return nil
		}(),
		DOB: user.Dob,
		Avatar: func() string {
			if user.Avatar.Valid && user.Avatar.String != "" {
				return app.File.GenerateSignImage(user.Avatar.String, userId, time.Now().Add(15*time.Minute))
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
		AvatarID:  user.Avatar.String,
		CanView:   permission,
	}

	if permission {
		userResp.Followers = &user.FollowersCount
		userResp.Following = &user.FollowingCount
	}
	return userResp
}
