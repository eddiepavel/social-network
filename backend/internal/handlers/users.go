package handlers

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_followers "social-network/pkg/db/queries/followers"
	db_posts "social-network/pkg/db/queries/posts"
	db_users "social-network/pkg/db/queries/users"
	"social-network/pkg/db/sqlite"
	"strconv"
	"strings"
	"time"
)

// GetUserProfile handles GET /api/users/:id
// Returns user profile respecting privacy settings
func GetUserProfile(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (for authentication check)
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Extract target user ID from URL path: /api/users/:id
		targetUserIDHex := r.PathValue("id")
		if targetUserIDHex == "" {
			utils.BadRequest(w, errors.New("user ID required"))
			return
		}

		targetUserID, err := helpers.GenerateFromString(targetUserIDHex)
		if err != nil {
			utils.BadRequest(w, err)
			return
		}

		// Fetch user profile

		user, err := sqlite.NewQuery(app.DB).Users.GetUserByIdWithCounts(r.Context(), targetUserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
			utils.Internal(w, errors.New("something went wrong"))
			return
		}

		var visitPermi bool
		isOwnProfile := false

		if bytes.Equal(user.UserID, userID) {
			visitPermi = true
			isOwnProfile = true
		} else {
			follower, err := sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(), db_followers.CheckIfUserFollowsParams{
				FollowerID: userID,
				FolloweeID: user.UserID,
			})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					visitPermi = false
				} else {
					utils.Internal(w, errors.New("Internal server error"))
					return
				}
			}

			if bytes.Equal(follower.FollowerID, userID) {
				visitPermi = true
			}
		}

		// Allow viewing any profile - access control for posts is handled separately
		// This allows users to see profiles and send follow requests to private accounts
		response := helpers.UserToResponseProfile(user, app, userID, visitPermi, isOwnProfile)

		utils.OK(w, response)
	}

}

// UpdateProfile handles PUT /api/users/profile
// Updates current user's profile (own profile only)
func UpdateProfile(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		user := helpers.FetchUser(app, userID, r.Context(), w)
		if user.UserID == nil {
			return // Error already handled in FetchUser
		}

		// Parse request body
		var req models.UpdateProfileRequest

		inputs := helpers.ValidateUpdateProfile.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusBadRequest, "400", "validation error", errValidation)
			return
		}

		if *req.Avatar != user.Avatar.String && *req.Avatar != "" {
			img := strings.Split(*req.Avatar, ".")

			if len(img) != 2 {
				app.Logger.Error("image length error")
				utils.BadRequest(w, errors.New("wrong payload"))
				return
			}

			if err := app.File.AssignImage(img[0]); err != nil {
				utils.Internal(w, errors.New("something went wrong"))
				return
			}

			if user.Avatar.Valid && user.Avatar.String != "" {

				img := strings.Split(user.Avatar.String, ".")

				if len(img) != 2 {
					utils.Internal(w, errors.New("something went wrong"))
					return
				}

				if err := app.File.RemoveImage(img[0]); err != nil {
					utils.Internal(w, errors.New("something went wrong"))
					return
				}
			}

		}

		updated, err := sqlite.NewQuery(app.DB).Users.UpdateUser(r.Context(), db_users.UpdateUserParams{
			FirstName: *req.FirstName,
			LastName:  *req.LastName,
			Nickname:  sql.NullString{Valid: true, String: *req.Nickname},
			AboutMe:   sql.NullString{Valid: true, String: *req.AboutMe},
			Avatar:    sql.NullString{Valid: true, String: *req.Avatar},
			UserID:    userID,
		})
		if err != nil {
			app.Logger.Error("Failed to update user", "error", err.Error())
			utils.Internal(w, err)
			return
		}

		// Return user response
		response := helpers.UserToResponse(updated)

		utils.OK(w, response)
	}
}

// UpdatePrivacy handles PUT /api/users/privacy
// Updates current user's privacy settings (own profile only)
func UpdatePrivacy(app *app.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Parse request body
		var req models.UpdatePrivacyRequest

		inputs := helpers.ValidatePrivacy.Build(r, app)

		ok, errorValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errorValidation)
			return
		}

		// Update privacy setting
		user, err := sqlite.NewQuery(app.DB).Users.UpdateUserPrivacy(r.Context(), db_users.UpdateUserPrivacyParams{IsPublic: req.IsPublic, UserID: userID})
		if err == sql.ErrNoRows {
			utils.NotFound(w)
			return
		}

		if err != nil {
			app.Logger.Error("Failed to fetch user", "error", err.Error())
			utils.Internal(w, err)
			return
		}

		// Batch update posts visibility when the profile privacy changes, except for "private" posts
		if !req.IsPublic {
			err = sqlite.NewQuery(app.DB).Posts.PostVisibilitySemiPrivateBatch(r.Context(), userID)
			if err != nil {
				app.Logger.Error("Failed to update posts visibility", "error", err.Error())
				utils.Internal(w, err)
			}
		} else {
			err = sqlite.NewQuery(app.DB).Posts.PostVisibilityPublicBatch(r.Context(), userID)
			if err != nil {
				app.Logger.Error("Failed to update posts visibility", "error", err.Error())
				utils.Internal(w, err)
			}
		}

		// Return updated user
		response := helpers.UserToResponse(user)

		utils.OK(w, response)
	}
}

// GetUserPosts handles GET /api/users/profile/{id}/posts
// Returns posts by a specific user respecting visibility rules
func GetUserPosts(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		targetUserIDHex := r.PathValue("id")
		if targetUserIDHex == "" {
			utils.BadRequest(w, errors.New("user ID required"))
			return
		}

		targetUserID, err := helpers.GenerateFromString(targetUserIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user ID format"))
			return
		}

		// Verify target user exists
		targetUser := helpers.FetchUser(app, targetUserID, r.Context(), w)
		if targetUser.UserID == nil {
			return
		}

		// Parse pagination parameters
		page := 1
		size := 10

		if pageParam := r.URL.Query().Get("page"); pageParam != "" {
			if parsedPage, err := strconv.Atoi(pageParam); err == nil && parsedPage > 0 {
				page = parsedPage
			}
		}

		if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
			if parsedSize, err := strconv.Atoi(sizeParam); err == nil && parsedSize > 0 {
				size = parsedSize
			}
		}

		offset := int64((page - 1) * size)
		limit := int64(size)

		posts, err := sqlite.NewQuery(app.DB).Posts.GetUserPosts(r.Context(), db_posts.GetUserPostsParams{
			AuthorID:   currentUserID, // for user_reacted check
			AuthorID_2: targetUserID,  // whose posts to fetch
			FollowerID: currentUserID, // for semi-private visibility check
			AuthorID_3: currentUserID, // for private visibility - check if viewing own posts
			UserID:     currentUserID, // for viewing_permissions check
			Limit:      limit,
			Offset:     offset,
		})

		if err != nil {
			app.Logger.Error("failed to get user posts", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		var userPosts []models.FeedPostResponse

		if len(posts) == 0 {
			utils.OK(w, userPosts)
			return
		}

		for _, post := range posts {
			postUuid, _ := helpers.GenerateFromBytes(post.PostID)
			authorUuid, _ := helpers.GenerateFromBytes(post.AuthorID)

			userPosts = append(userPosts, models.FeedPostResponse{
				PostID:   postUuid,
				AuthorID: authorUuid,
				Content:  post.Content,
				ImageID:  &post.ImageID.String,
				ImageUrl: func() string {
					if post.ImageID.Valid {
						path := app.File.GenerateSignImage(post.FileName.String, currentUserID, time.Now().Add(15*time.Minute))
						return path
					}
					return ""
				}(),
				AuthorFirstName: post.FirstName,
				AuthorLastName:  post.LastName,
				AuthorAvatar: func() *string {
					if post.Avatar.Valid && post.Avatar.String != "" {
						path := app.File.GenerateSignImage(post.Avatar.String, currentUserID, time.Now().Add(15*time.Minute))
						return &path
					}
					return nil
				}(),
				Visibility: post.Visibility,
				CreatedAt: func() time.Time {
					if post.CreatedAt.Valid {
						return post.CreatedAt.Time
					}
					return time.Time{}
				}(),
				UserReacted:   post.UserReacted != 0,
				ReactionCount: post.ReactionCount,
				CommentCount:  post.CommentCount,
			})
		}

		utils.OK(w, userPosts)
	}
}

func SearchUsers(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user_id, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "no user id found")
			return
		}

		searchParams := strings.TrimSpace(r.URL.Query().Get("name"))

		users := []models.UserResponse{}

		searchUsers, err := sqlite.NewQuery(app.DB).Users.QueryUsers(r.Context(), db_users.QueryUsersParams{
			Column1: sql.NullString{Valid: true, String: searchParams},
			Column2: sql.NullString{Valid: true, String: searchParams},
			Column3: sql.NullString{Valid: true, String: searchParams},
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.OK(w, users)
				return
			}
			utils.Internal(w, errors.New("database error"))
			return
		}

		for _, user := range searchUsers {
			userID, _ := helpers.GenerateFromBytes(user.UserID)
			if bytes.Equal(user.UserID, user_id) {
				continue
			}
			users = append(users, models.UserResponse{
				UserID:    userID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Nickname:  user.Nickname.String,
				Avatar: func() string {
					if user.Avatar.Valid && user.Avatar.String != "" {
						img := app.File.GenerateSignImage(user.Avatar.String, user_id, time.Now().Add(15*time.Minute))
						return img
					}
					return ""
				}(),
			})
		}

		utils.OK(w, users)
	}
}
