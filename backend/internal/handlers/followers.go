package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/utils"
	db_followers "social-network/pkg/db/queries/followers"
	db_requests "social-network/pkg/db/queries/followers/requests"
	"social-network/pkg/db/sqlite"
	"strconv"
	"time"
)

// FollowerResponse represents a follower/following user
type FollowerResponse struct {
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Avatar    *string   `json:"avatar"`
	Nickname  *string   `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
}

type FollowRequestsResponse struct {
	ID           int64     `json:"id"`
	FollowerID   string    `json:"follower_id"`
	FollowerName string    `json:"follower_name"`
	CreatedAt    time.Time `json:"created_at"`
}

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

		userID, err := hex.DecodeString(userIDHex)
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

		if user.IsPublic.Bool {
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
			request, err := sqlite.NewQuery(app.DB).FollowRequests.GetFollowRequest(r.Context(),
				db_requests.GetFollowRequestParams{
					FollowerID: currentUserID,
					FolloweeID: user.UserID,
				})
			if errors.Is(err, sql.ErrNoRows) {
				err := sqlite.NewQuery(app.DB).FollowRequests.CreateFollowRequest(r.Context(),
					db_requests.CreateFollowRequestParams{
						FollowerID: currentUserID,
						FolloweeID: user.UserID,
					})
				if err != nil {
					utils.Internal(w, err)
					return
				}
				utils.OK(w, "Follow requested successfully")
				return
			} else {
				if err != nil {
					utils.Internal(w, err)
					return
				}
				err := sqlite.NewQuery(app.DB).FollowRequests.DeleteFollowRequest(r.Context(), request.ID)
				if err != nil {
					utils.Internal(w, err)
					return
				}
				utils.OK(w, "Follow request cancelled successfully")
				return
			}
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

		request, err := sqlite.NewQuery(app.DB).FollowRequests.GetFollowRequestByID(r.Context(),
			db_requests.GetFollowRequestByIDParams{
				ID:         int64(requestRaw),
				FolloweeID: currentUserID,
			})
		if errors.Is(err, sql.ErrNoRows) {
			utils.NotFound(w)
		}
		if err != nil {
			utils.Internal(w, err)
			return
		}

		var req map[string]interface{}

		err = json.NewDecoder(r.Body).Decode(&req)

		if req["status"].(string) == "accepted" {
			err = sqlite.NewQuery(app.DB).FollowRequests.AcceptFollowRequest(r.Context(), request.ID)
			if err != nil {
				utils.Internal(w, err)
				return
			}

			utils.OK(w, "Follow request accepted")
		} else if req["status"].(string) == "rejected" {
			err = sqlite.NewQuery(app.DB).FollowRequests.DeleteFollowRequest(r.Context(), request.ID)
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

		userID, err := hex.DecodeString(userIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user ID format"))
			return
		}

		user := helpers.FetchUser(app, userID, r.Context(), w)
		if user.UserID == nil {
			return
		}

		if !(string(user.UserID) == string(currentUserID) || user.IsPublic.Bool) {
			_, err = sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(),
				db_followers.CheckIfUserFollowsParams{
					FollowerID: currentUserID,
					FolloweeID: user.UserID,
				})
			if errors.Is(err, sql.ErrNoRows) {
				utils.OK(w, []FollowerResponse{})
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

		var followersList []FollowerResponse

		for _, follower := range followers {
			user, err := sqlite.NewQuery(app.DB).Users.GetUserById(r.Context(), follower.FollowerID)
			if err != nil {
				utils.Internal(w, err)
				return
			}
			followersList = append(followersList, FollowerResponse{
				UserID:    hex.EncodeToString(user.UserID),
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Avatar: func() *string {
					if user.Avatar.Valid {
						return &user.Avatar.String
					}
					return nil
				}(),
				Nickname: &user.Nickname,
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

		userID, err := hex.DecodeString(userIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user ID format"))
			return
		}

		user := helpers.FetchUser(app, userID, r.Context(), w)
		if user.UserID == nil {
			return
		}

		if !(string(user.UserID) == string(currentUserID) || user.IsPublic.Bool) {
			_, err = sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(),
				db_followers.CheckIfUserFollowsParams{
					FollowerID: currentUserID,
					FolloweeID: user.UserID,
				})
			if errors.Is(err, sql.ErrNoRows) {
				utils.OK(w, []FollowerResponse{})
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

		var followersList []FollowerResponse

		for _, follower := range followers {
			user, err := sqlite.NewQuery(app.DB).Users.GetUserById(r.Context(), follower.FollowerID)
			if err != nil {
				utils.Internal(w, err)
				return
			}
			followersList = append(followersList, FollowerResponse{
				UserID:    hex.EncodeToString(user.UserID),
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Avatar: func() *string {
					if user.Avatar.Valid {
						return &user.Avatar.String
					}
					return nil
				}(),
				Nickname: &user.Nickname,
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
		requests, err := sqlite.NewQuery(app.DB).FollowRequests.GetFollowRequests(r.Context(), currentUserID)
		if errors.Is(err, sql.ErrNoRows) {
			utils.OK(w, []FollowRequestsResponse{})
			return
		}
		if err != nil {
			app.Logger.Error("Failed to fetch follow requests", "error", err)
			utils.Internal(w, err)
			return
		}

		var response []FollowRequestsResponse

		for _, request := range requests {
			user, _ := sqlite.NewQuery(app.DB).Users.GetUserById(r.Context(), request.FollowerID)
			response = append(response, FollowRequestsResponse{
				ID:           request.ID,
				FollowerID:   hex.EncodeToString(request.FollowerID),
				FollowerName: user.FirstName + " " + user.LastName,
				CreatedAt:    request.CreatedAt,
			})
		}

		utils.OK(w, response)
	}
}
