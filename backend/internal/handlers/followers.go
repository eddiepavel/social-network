package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/constants"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_followers "social-network/pkg/db/queries/followers"
	db_posts "social-network/pkg/db/queries/posts"
	"social-network/pkg/db/sqlite"
	"strconv"
	"time"
)

// FollowUser handles POST /api/follow/:userId
// Sends follow request or auto-follows if public profile
func FollowUser(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract target user ID from URL
		userIDHex := r.PathValue("userId")
		if userIDHex == "" {
			utils.BadRequest(w, errors.New("user ID required"))
			return
		}

		userID, err := helpers.GenerateFromString(userIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user ID format"))
			return
		}

		// Can't follow yourself
		if string(currentUserID) == string(userID) {
			utils.BadRequest(w, errors.New("cannot follow yourself"))
			return
		}

		user := helpers.FetchUser(app, userID, r.Context(), w)
		if user.UserID == nil {
			return
		}

		if user.IsPublic {
			_, err := sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(),
				db_followers.CheckIfUserFollowsParams{
					FollowerID: currentUserID,
					FolloweeID: user.UserID,
				})
			if errors.Is(err, sql.ErrNoRows) {
				err := sqlite.NewQuery(app.DB).Followers.InsertFollower(r.Context(),
					db_followers.InsertFollowerParams{
						FollowerID: currentUserID,
						FolloweeID: user.UserID,
					})
				if err != nil {
					utils.Internal(w, err)
					return
				}

				utils.OK(w, "Followed successfully")
				return
			} else {
				if err != nil {
					utils.Internal(w, err)
					return
				}
				err := sqlite.NewQuery(app.DB).Followers.DeleteFollower(r.Context(),
					db_followers.DeleteFollowerParams{
						FollowerID: currentUserID,
						FolloweeID: user.UserID,
					})
				if err != nil {
					utils.Internal(w, err)
					return
				}
				// in case of unfollowing, remove viewing permissions from posts
				err = sqlite.NewQuery(app.DB).Posts.RemoveViewingPermissionUserIDBatch(r.Context(),
					db_posts.RemoveViewingPermissionUserIDBatchParams{
						UserID:   currentUserID,
						AuthorID: user.UserID,
					})
				utils.OK(w, "Unfollowed successfully")
				return
			}
		}

		_, err = sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(),
			db_followers.CheckIfUserFollowsParams{
				FollowerID: currentUserID,
				FolloweeID: user.UserID,
			})
		if errors.Is(err, sql.ErrNoRows) {
			request, err := sqlite.NewQuery(app.DB).Followers.GetFollowRequest(r.Context(),
				db_followers.GetFollowRequestParams{
					FollowerID: currentUserID,
					FolloweeID: user.UserID,
				})
			if errors.Is(err, sql.ErrNoRows) {
				err := sqlite.NewQuery(app.DB).Followers.CreateFollowRequest(r.Context(),
					db_followers.CreateFollowRequestParams{
						FollowerID: currentUserID,
						FolloweeID: user.UserID,
					})
				if err != nil {
					utils.Internal(w, err)
					return
				}

				// Create notification for follow request
				err = helpers.CreateNotification(app, user.UserID, constants.NotificationFollowRequest, currentUserID, nil, nil, nil)
				if err != nil {
					app.Logger.Error("failed to create follow request notification", "err", err)
					// Don't fail the request if notification fails
				}

				utils.OK(w, "Follow requested successfully")
				return
			} else {
				if err != nil {
					utils.Internal(w, err)
					return
				}
				err := sqlite.NewQuery(app.DB).Followers.DeleteFollowRequest(r.Context(), request.ID)
				if err != nil {
					utils.Internal(w, err)
					return
				}
				utils.OK(w, "Follow request cancelled successfully")
				return
			}
		} else {
			if err != nil {
				utils.Internal(w, err)
				return
			}
			err := sqlite.NewQuery(app.DB).Followers.DeleteFollower(r.Context(),
				db_followers.DeleteFollowerParams{
					FollowerID: currentUserID,
					FolloweeID: user.UserID,
				})
			if err != nil {
				utils.Internal(w, err)
				return
			}
			// in case of unfollowing, remove viewing permissions from posts
			err = sqlite.NewQuery(app.DB).Posts.RemoveViewingPermissionUserIDBatch(r.Context(),
				db_posts.RemoveViewingPermissionUserIDBatchParams{
					UserID:   currentUserID,
					AuthorID: user.UserID,
				})
			utils.OK(w, "Unfollowed successfully")
			return
		}

	}
}

// AcceptFollowRequest handles POST /api/follow/accept/:followerId
func UpdateFollowRequest(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (the followee)
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		requestRaw, err := strconv.Atoi(r.PathValue("requestId"))
		if err != nil {
			utils.BadRequest(w, errors.New("request ID required"))
			return
		}

		request, err := sqlite.NewQuery(app.DB).Followers.GetFollowRequestByID(r.Context(),
			db_followers.GetFollowRequestByIDParams{
				ID:         int64(requestRaw),
				FolloweeID: currentUserID,
			})
		if errors.Is(err, sql.ErrNoRows) {
			utils.NotFound(w)
			return
		}
		if err != nil {
			utils.Internal(w, err)
			return
		}

		var req map[string]interface{}

		err = json.NewDecoder(r.Body).Decode(&req)

		if err != nil {
			utils.BadRequest(w, errors.New("bad request"))
			return
		}

		status, ok := req["status"].(string)

		if !ok {
			utils.BadRequest(w, errors.New("bad request"))
			return
		}

		if status == "accepted" {
			err = sqlite.NewQuery(app.DB).Followers.DeleteFollowRequest(r.Context(), request.ID)
			if err != nil {
				utils.Internal(w, err)
				return
			}

			err = sqlite.NewQuery(app.DB).Followers.InsertFollower(r.Context(), db_followers.InsertFollowerParams{
				FollowerID: request.FollowerID,
				FolloweeID: request.FolloweeID,
			})

			if err != nil {
				utils.Internal(w, errors.New("internal server error"))
				return
			}

			// Create notification for follow accepted
			err = helpers.CreateNotification(app, request.FollowerID, constants.NotificationFollowAccepted, currentUserID, nil, nil, nil)
			if err != nil {
				app.Logger.Error("failed to create follow accepted notification", "err", err)
				// Don't fail the request if notification fails
			}

			utils.OK(w, "Follow request accepted")
		} else if status == "rejected" {
			err = sqlite.NewQuery(app.DB).Followers.DeleteFollowRequest(r.Context(), request.ID)
			if err != nil {
				utils.Internal(w, err)
				return
			}

			utils.OK(w, "Follow request rejected")
		} else {
			utils.BadRequest(w, errors.New("invalid status"))
		}
	}
}

// GetFollowers handles GET /api/followers/:userId
func GetFollowers(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Extract target user ID from URL
		userIDHex := r.PathValue("userId")
		if userIDHex == "" {
			utils.BadRequest(w, errors.New("user ID required"))
			return
		}

		userID, err := helpers.GenerateFromString(userIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user ID format"))
			return
		}

		user := helpers.FetchUser(app, userID, r.Context(), w)
		if user.UserID == nil {
			return
		}

		if !(string(user.UserID) == string(currentUserID) || user.IsPublic) {
			_, err = sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(),
				db_followers.CheckIfUserFollowsParams{
					FollowerID: currentUserID,
					FolloweeID: user.UserID,
				})
			if errors.Is(err, sql.ErrNoRows) {
				utils.OK(w, []models.FollowerResponse{})
			}
			if err != nil {
				utils.Internal(w, err)
				return
			}
		}

		followers, err := sqlite.NewQuery(app.DB).Followers.GetFollowers(r.Context(), user.UserID)
		if err != nil {
			utils.Internal(w, err)
			return
		}

		var followersList []models.FollowerResponse

		for _, follower := range followers {
			user, err := sqlite.NewQuery(app.DB).Users.GetUserById(r.Context(), follower.FollowerID)
			if err != nil {
				utils.Internal(w, err)
				return
			}
			getUuid, _ := helpers.GenerateFromBytes(user.UserID)
			followersList = append(followersList, models.FollowerResponse{
				UserID:    getUuid,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Avatar: func() *string {
					if user.Avatar.Valid && user.Avatar.String != "" {
						img := app.File.GenerateSignImage(user.Avatar.String, currentUserID, time.Now().Add(15*time.Minute))
						return &img
					}
					return nil
				}(),
				Nickname: func() *string {
					if user.Nickname.Valid {
						return &user.Nickname.String
					}
					return nil
				}(),
				CreatedAt: func() time.Time {
					if follower.CreatedAt.Valid {
						return follower.CreatedAt.Time
					}
					return time.Time{}
				}(),
			})
		}

		utils.OK(w, followersList)
	}
}

// GetFollowing handles GET /api/following/:userId
func GetFollowing(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Extract target user ID from URL
		userIDHex := r.PathValue("userId")
		if userIDHex == "" {
			utils.BadRequest(w, errors.New("user ID required"))
			return
		}

		userID, err := helpers.GenerateFromString(userIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user ID format"))
			return
		}

		user := helpers.FetchUser(app, userID, r.Context(), w)
		if user.UserID == nil {
			return
		}

		if !(string(user.UserID) == string(currentUserID) || user.IsPublic) {
			_, err = sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(),
				db_followers.CheckIfUserFollowsParams{
					FollowerID: currentUserID,
					FolloweeID: user.UserID,
				})
			if errors.Is(err, sql.ErrNoRows) {
				utils.OK(w, []models.FollowerResponse{})
			}
			if err != nil {
				utils.Internal(w, err)
				return
			}
		}

		followers, err := sqlite.NewQuery(app.DB).Followers.GetFollowees(r.Context(), user.UserID)
		if err != nil {
			utils.Internal(w, err)
			return
		}

		var followersList []models.FollowerResponse

		for _, follower := range followers {
			user, err := sqlite.NewQuery(app.DB).Users.GetUserById(r.Context(), follower.FolloweeID)
			if err != nil {
				utils.Internal(w, err)
				return
			}
			getUuid, _ := helpers.GenerateFromBytes(user.UserID)
			followersList = append(followersList, models.FollowerResponse{
				UserID:    getUuid,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Avatar: func() *string {
					if user.Avatar.Valid && user.Avatar.String != "" {
						img := app.File.GenerateSignImage(user.Avatar.String, currentUserID, time.Now().Add(15*time.Minute))
						return &img
					}
					return nil
				}(),
				Nickname: func() *string {
					if user.Nickname.Valid {
						return &user.Nickname.String
					}
					return nil
				}(),
				CreatedAt: func() time.Time {
					if follower.CreatedAt.Valid {
						return follower.CreatedAt.Time
					}
					return time.Time{}
				}(),
			})
		}

		utils.OK(w, followersList)
	}
}

// GetFollowStatus handles GET /api/followers/status/{userId}
// Returns the follow relationship status between current user and target user
func GetFollowStatus(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		userIDHex := r.PathValue("userId")
		if userIDHex == "" {
			utils.BadRequest(w, errors.New("user ID required"))
			return
		}

		userID, err := helpers.GenerateFromString(userIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user ID format"))
			return
		}

		// Check if it's the same user
		if string(currentUserID) == string(userID) {
			utils.OK(w, models.FollowStatusResponse{Status: "self"})
			return
		}

		// Check if already following
		_, err = sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(),
			db_followers.CheckIfUserFollowsParams{
				FollowerID: currentUserID,
				FolloweeID: userID,
			})
		if err == nil {
			utils.OK(w, models.FollowStatusResponse{Status: "following"})
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			utils.Internal(w, err)
			return
		}

		// Check for pending follow request
		_, err = sqlite.NewQuery(app.DB).Followers.CheckPendingFollowRequest(r.Context(),
			db_followers.CheckPendingFollowRequestParams{
				FollowerID: currentUserID,
				FolloweeID: userID,
			})
		if err == nil {
			utils.OK(w, models.FollowStatusResponse{Status: "requested"})
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			utils.Internal(w, err)
			return
		}

		utils.OK(w, models.FollowStatusResponse{Status: "none"})
	}
}

// GetFollowRequests handles GET /api/follow/requests
// Returns pending follow requests for the current user
func GetFollowRequests(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Get pending follow requests
		requests, err := sqlite.NewQuery(app.DB).Followers.GetFollowRequests(r.Context(), currentUserID)
		if errors.Is(err, sql.ErrNoRows) {
			utils.OK(w, []models.FollowRequestsResponse{})
			return
		}
		if err != nil {
			app.Logger.Error("Failed to fetch follow requests", "error", err)
			utils.Internal(w, err)
			return
		}

		var response []models.FollowRequestsResponse

		for _, request := range requests {
			user, err := sqlite.NewQuery(app.DB).Users.GetUserById(r.Context(), request.FollowerID)
			if err != nil {
				app.Logger.Error("Failed to fetch user for follow request", "error", err)
				continue
			}

			followerUUID, _ := helpers.GenerateFromBytes(request.FollowerID)
			requestID := strconv.FormatInt(request.ID, 10)

			response = append(response, models.FollowRequestsResponse{
				RequestID: requestID,
				UserID:    followerUUID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Nickname: func() *string {
					if user.Nickname.Valid {
						return &user.Nickname.String
					}
					return nil
				}(),
				Avatar: func() *string {
					if user.Avatar.Valid && user.Avatar.String != "" {
						img := app.File.GenerateSignImage(user.Avatar.String, currentUserID, time.Now().Add(15*time.Minute))
						return &img
					}
					return nil
				}(),
				CreatedAt: request.CreatedAt.Format(time.RFC3339),
			})
		}

		utils.OK(w, response)
	}
}
