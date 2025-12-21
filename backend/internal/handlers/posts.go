package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_posts "social-network/pkg/db/queries/posts"
	"social-network/pkg/db/sqlite"
	"time"

	"github.com/google/uuid"
)

// --------------------- GENERAL POSTS HANDLERS --------------------

func CreatePost(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req models.CreatePostRequest

		inputs := helpers.ValidatePost.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		// Generate UUID for post
		postID, err := uuid.New().MarshalBinary()
		if err != nil {
			app.Logger.Error("failed uuid", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		var image sql.NullString

		if req.ImageID != "" {
			image = sql.NullString{String: req.ImageID, Valid: true}
		} else {
			image = sql.NullString{String: "", Valid: false}
		}

		post, err := sqlite.NewQuery(app.DB).Posts.CreatePost(r.Context(), db_posts.CreatePostParams{
			PostID:     postID,
			AuthorID:   currentUserID,
			Content:    req.Content,
			Visibility: req.Visibility,
			ImageID:    image,
		})

		if err != nil {
			app.Logger.Error("failed to create post", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		postTime := post.CreatedAt.Time

		postResponse := models.PostResponse{
			PostID:     post.PostID,
			Content:    post.Content,
			AuthorID:   post.AuthorID,
			Visibility: post.Visibility,
			CreatedAt:  postTime,
			ImageID:    image.String,
		}

		utils.OK(w, postResponse)
	}
}

func GetFeedPosts(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// TODO: Add query parameters for limit and offset with default values
		// For now using hardcoded values
		limit := int64(50)
		offset := int64(0)

		posts, err := sqlite.NewQuery(app.DB).Posts.GetPostsForFeed(r.Context(), db_posts.GetPostsForFeedParams{
			FollowerID: currentUserID,
			UserID:     currentUserID,
			Limit:      limit,
			Offset:     offset,
		})

		if err != nil {
			app.Logger.Error("failed to get feed posts", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		var feedPosts []models.FeedPostResponse
		for _, post := range posts {
			feedPosts = append(feedPosts, models.FeedPostResponse{
				PostID:   post.PostID,
				AuthorID: post.AuthorID,
				Content:  post.Content,
				ImageID: func() string {
					if post.ImageID.Valid {
						return post.ImageID.String
					}
					return ""
				}(),
				Visibility: post.Visibility,
				CreatedAt: func() time.Time {
					if post.CreatedAt.Valid {
						return post.CreatedAt.Time
					}
					return time.Time{}
				}(),
				ReactionCount: post.ReactionCount,
				CommentCount:  post.CommentCount,
			})
		}

		utils.OK(w, feedPosts)
	}
}

func GetPostWithCommentsReactions(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		postIDHex := r.PathValue("postId")
		if postIDHex == "" {
			utils.BadRequest(w, errors.New("post ID required"))
			return
		}

		postID, err := helpers.GenerateFromString(postIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid post ID format"))
			return
		}

		postBasicInfo := helpers.FetchPostBasicInfo(app, postID, r.Context(), w)
		if postBasicInfo.PostID == nil {
			return
		}

		// Check if user has permission to view this post
		canView, err := helpers.CanViewPost(currentUserID, postBasicInfo.PostID, postBasicInfo.AuthorID, postBasicInfo.Visibility, app, r)
		if err != nil {
			app.Logger.Error("failed to check post permissions", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if !canView {
			utils.Forbidden(w)
			return
		}

		// Now fetch the full post data with reactions and comments
		post := helpers.FetchPost(app, postID, r.Context(), w)
		if post.PostID == nil {
			return
		}

		// Parse reactions JSON
		var reactions []models.Reaction
		if post.Reactions != nil {
			reactionsJSON, _ := post.Reactions.([]byte)
			json.Unmarshal(reactionsJSON, &reactions)
		}

		// Parse comments JSON
		var comments []models.Comment
		if post.Comments != nil {
			commentsJSON, _ := post.Comments.([]byte)
			json.Unmarshal(commentsJSON, &comments)
		}

		response := models.PostWithCommentsReactionsResponse{
			PostID:   post.PostID,
			AuthorID: post.AuthorID,
			Content:  post.Content,
			ImageID: func() string {
				if post.ImageID.Valid {
					return post.ImageID.String
				}
				return ""
			}(),
			Visibility: post.Visibility,
			CreatedAt: func() time.Time {
				if post.CreatedAt.Valid {
					return post.CreatedAt.Time
				}
				return time.Time{}
			}(),
			Reactions: reactions,
			Comments:  comments,
		}

		utils.OK(w, response)
	}
}

func EditPost(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		postIDHex := r.PathValue("postId")
		if postIDHex == "" {
			utils.BadRequest(w, errors.New("post ID required"))
			return
		}

		postID, err := helpers.GenerateFromString(postIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid post ID format"))
			return
		}

		var req models.UpdatePostRequest

		inputs := helpers.ValidateUpdatePost.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		var image sql.NullString
		if req.ImageID != "" {
			image = sql.NullString{String: req.ImageID, Valid: true}
		} else {
			image = sql.NullString{String: "", Valid: false}
		}

		err = sqlite.NewQuery(app.DB).Posts.UpdatePost(r.Context(), db_posts.UpdatePostParams{
			Content:  req.Content,
			ImageID:  image,
			PostID:   postID,
			AuthorID: currentUserID,
		})

		if err != nil {
			app.Logger.Error("failed to update post", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, "Post updated successfully")
	}
}

func DeletePost(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		postIDHex := r.PathValue("postId")
		if postIDHex == "" {
			utils.BadRequest(w, errors.New("post ID required"))
			return
		}

		postID, err := helpers.GenerateFromString(postIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid post ID format"))
			return
		}

		err = sqlite.NewQuery(app.DB).Posts.DeletePost(r.Context(), db_posts.DeletePostParams{
			PostID:   postID,
			AuthorID: currentUserID,
		})

		if err != nil {
			app.Logger.Error("failed to delete post", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, "Post deleted successfully")
	}
}

// -------------------- POSTS PRIVACY HANDLERS ---------------------

func UpdatePostVisibility(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		postIDHex := r.PathValue("postId")
		if postIDHex == "" {
			utils.BadRequest(w, errors.New("post ID required"))
			return
		}

		postID, err := helpers.GenerateFromString(postIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid post ID format"))
			return
		}

		var req models.UpdateVisibilityRequest

		inputs := helpers.ValidateUpdatePostVisibility.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		err = sqlite.NewQuery(app.DB).Posts.EditPostVisibility(r.Context(), db_posts.EditPostVisibilityParams{
			Visibility: req.Visibility,
			PostID:     postID,
			AuthorID:   currentUserID,
		})

		if err != nil {
			app.Logger.Error("failed to update post visibility", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, "Post visibility updated successfully")
	}
}

func AddUserToPrivatePostList(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		postIDHex := r.PathValue("postId")
		if postIDHex == "" {
			utils.BadRequest(w, errors.New("post ID required"))
			return
		}

		postID, err := helpers.GenerateFromString(postIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid post ID format"))
			return
		}

		var req models.AddUserToPrivatePostRequest

		inputs := helpers.ValidateAddUserToPrivatePost.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		targetUserID, err := helpers.GenerateFromString(req.UserID)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user ID format"))
			return
		}

		// Fetch post basic info for ownership and visibility check
		post := helpers.FetchPostBasicInfo(app, postID, r.Context(), w)
		if post.PostID == nil {
			return
		}

		// Verify the current user is the author of the post
		if string(post.AuthorID) != string(currentUserID) {
			utils.Forbidden(w)
			return
		}

		// Verify the post is private
		if post.Visibility != "private" {
			utils.BadRequest(w, errors.New("post is not private"))
			return
		}

		postIDString, _ := helpers.GenerateFromBytes(postID)

		rowsAffected, err := sqlite.NewQuery(app.DB).Posts.AddPrivatePostViewingPermission(r.Context(), db_posts.AddPrivatePostViewingPermissionParams{
			UserID:     targetUserID,
			PostID:     postIDString,
			PostID_2:   postID,
			FollowerID: targetUserID,
		})

		if err != nil {
			app.Logger.Error("failed to add viewing permission", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if rowsAffected == 0 {
			utils.BadRequest(w, errors.New("user must be following you to view private post"))
			return
		}

		utils.OK(w, "User added to private post viewing list")
	}
}

func RemoveUserFromPrivatePostList(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		postIDHex := r.PathValue("postId")
		if postIDHex == "" {
			utils.BadRequest(w, errors.New("post ID required"))
			return
		}

		postID, err := helpers.GenerateFromString(postIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid post ID format"))
			return
		}

		var req models.RemoveUserFromPrivatePostRequest

		inputs := helpers.ValidateAddUserToPrivatePost.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		targetUserID, err := helpers.GenerateFromString(req.UserID)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user ID format"))
			return
		}

		// Fetch post basic info for ownership check
		post := helpers.FetchPostBasicInfo(app, postID, r.Context(), w)
		if post.PostID == nil {
			return
		}

		// Verify the current user is the author of the post
		if string(post.AuthorID) != string(currentUserID) {
			utils.Forbidden(w)
			return
		}

		postIDString, _ := helpers.GenerateFromBytes(postID)

		err = sqlite.NewQuery(app.DB).Posts.RemovePrivatePostViewingPermission(r.Context(), db_posts.RemovePrivatePostViewingPermissionParams{
			UserID: targetUserID,
			PostID: postIDString,
		})

		if err != nil {
			app.Logger.Error("failed to remove viewing permission", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, "User removed from private post viewing list")
	}
}
