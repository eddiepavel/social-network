package handlers

import (
	"database/sql"
	"errors"
	"fmt"
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

		user := helpers.FetchUser(app, currentUserID, r.Context(), w)

		if !user.IsPublic && req.Visibility == "public" {
			utils.BadRequest(w, errors.New("cannot post publicly while having a private profile"))
			return
		} else if user.IsPublic && req.Visibility == "semi-private" {
			utils.BadRequest(w, errors.New("cannot post semi-privately while having a public profile"))
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

		// For private posts, add viewing permissions for selected followers
		if req.Visibility == "private" && len(req.AllowedUsers) > 0 {
			query := sqlite.NewQuery(app.DB)
			for _, userIDStr := range req.AllowedUsers {
				allowedUserID, err := helpers.GenerateFromString(userIDStr)
				if err != nil {
					app.Logger.Warn("invalid allowed_user ID", "id", userIDStr, "error", err)
					continue
				}
				// Verify the user is actually a follower
				_, err = query.Followers.CheckIfUserFollows(r.Context(), db_followers.CheckIfUserFollowsParams{
					FollowerID: allowedUserID,
					FolloweeID: currentUserID,
				})
				if err != nil {
					app.Logger.Warn("allowed_user is not a follower, skipping", "id", userIDStr)
					continue
				}
				// Add viewing permission
				err = query.Posts.AddViewingPermission(r.Context(), db_posts.AddViewingPermissionParams{
					UserID: allowedUserID,
					PostID: postID,
				})
				if err != nil {
					app.Logger.Warn("failed to add viewing permission", "user", userIDStr, "error", err)
				}
			}
		}

		postTime := post.CreatedAt.Time

		app.File.AssignImage(post.ImageID.String)

		postIDString, _ := helpers.GenerateFromBytes(postID)

		postResponse := models.PostResponse{
			PostID:     postIDString,
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

		// Parse query parameters for pagination
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

		// Calculate offset
		offset := int64((page - 1) * size)
		limit := int64(size)

		totalCount, err := sqlite.NewQuery(app.DB).Posts.GetFeedPostsCount(r.Context(), db_posts.GetFeedPostsCountParams{
			AuthorID:   currentUserID,
			FollowerID: currentUserID,
			UserID:     currentUserID,
			UserID_2:   currentUserID,
		})

		if err != nil {
			app.Logger.Error("failed to get feed posts count", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if totalCount == 0 {
			utils.OK(w, []models.FeedPostResponse{})
			return
		}

		// Get posts for current page
		posts, err := sqlite.NewQuery(app.DB).Posts.GetPostsForFeed(r.Context(), db_posts.GetPostsForFeedParams{
			AuthorID:   currentUserID,
			AuthorID_2: currentUserID,
			FollowerID: currentUserID,
			UserID:     currentUserID,
			UserID_2:   currentUserID,
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
			postUuid, _ := helpers.GenerateFromBytes(post.PostID)
			authorUuid, _ := helpers.GenerateFromBytes(post.AuthorID)

			feedPosts = append(feedPosts, models.FeedPostResponse{
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
				AuthorNickname: func() *string {
					if post.Nickname.Valid {
						return &post.Nickname.String
					}
					return nil
				}(),
				AuthorAvatar: func() *string {
					if post.Avatar.Valid && post.Avatar.String != "" {
						img := app.File.GenerateSignImage(post.Avatar.String, currentUserID, time.Now().Add(15*time.Minute))
						return &img
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

		// Calculate total pages
		totalPages := int(totalCount) / size
		if int(totalCount)%size != 0 {
			totalPages++
		}

		// Send response with pagination metadata
		utils.OK(w, utils.WithPagination(feedPosts, utils.Pagination{
			Page:       page,
			Size:       size,
			Current:    len(feedPosts),
			TotalItems: int(totalCount),
			TotalPages: totalPages,
		}))
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
		canView, err := helpers.CanViewPost(currentUserID, postBasicInfo.PostID, postBasicInfo.AuthorID, postBasicInfo.GroupID, postBasicInfo.Visibility, app, r)
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

		reactions, err := sqlite.NewQuery(app.DB).Posts.GetPostReactions(r.Context(), postID)
		if err != nil {
			app.Logger.Error("failed to get post reactions", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		postComments, err := sqlite.NewQuery(app.DB).Posts.GetPostComments(r.Context(), db_posts.GetPostCommentsParams{
			PostID:   postID,
			AuthorID: currentUserID,
		})
		if err != nil {
			app.Logger.Error("failed to get post comments", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		hasUserReacted, err := sqlite.NewQuery(app.DB).Posts.HasUserReacted(r.Context(), db_posts.HasUserReactedParams{
			AuthorID:   currentUserID,
			TargetType: "post",
			TargetID:   postID,
		})

		var comments []models.Comment
		var commentCount int64
		for _, comment := range postComments {
			commentCount++
			commentAuthorID, _ := helpers.GenerateFromBytes(comment.AuthorID)
			commentID, _ := helpers.GenerateFromBytes(comment.CommentID)
			var parentID string
			if comment.ParentCommentID != nil {
				parentID, _ = helpers.GenerateFromBytes(comment.ParentCommentID)
			}
			comments = append(comments, models.Comment{
				CommentID:       commentID,
				AuthorID:        commentAuthorID,
				AuthorFirstName: comment.FirstName,
				AuthorLastName:  comment.LastName,
				AuthorNickname: func() *string {
					if comment.Nickname.Valid {
						return &comment.Nickname.String
					}
					return nil
				}(),
				AuthorAvatar: func() *string {
					if comment.Avatar.Valid {
						return &comment.Avatar.String
					}
					return nil
				}(),
				Content:         comment.Content,
				ParentCommentID: &parentID,
				ImageID:         &comment.ImageID.String,
				ImageUrl: func() string {
					if comment.ImageID.Valid && comment.FileName.Valid {
						path := app.File.GenerateSignImage(comment.FileName.String, currentUserID, time.Now().Add(15*time.Minute))
						return path
					}
					return ""
				}(),
				CreatedAt: func() time.Time {
					if comment.CreatedAt.Valid {
						return comment.CreatedAt.Time
					}
					return time.Time{}
				}(),
				Reactions:   int(comment.ReactionCount),
				UserReacted: comment.UserReacted != 0,
			})
		}

		authorUuid, _ := helpers.GenerateFromBytes(post.AuthorID)
		response := models.PostWithCommentsReactionsResponse{
			PostID:  postIDHex,
			Content: post.Content,
			ImageID: func() *string {
				if post.ImageID.Valid {
					return &post.ImageID.String
				}
				return nil
			}(),
			ImageUrl: func() string {
				if post.ImageID.Valid && post.FileName.Valid {
					path := app.File.GenerateSignImage(post.FileName.String, currentUserID, time.Now().Add(15*time.Minute))
					return path
				}
				return ""
			}(),
			AuthorID:        authorUuid,
			AuthorFirstName: post.FirstName,
			AuthorLastName:  post.LastName,
			AuthorNickname: func() *string {
				if post.Nickname.Valid {
					return &post.Nickname.String
				}
				return nil
			}(),
			AuthorAvatar: func() *string {
				if post.Avatar.Valid && post.Avatar.String != "" {
					img := app.File.GenerateSignImage(post.Avatar.String, currentUserID, time.Now().Add(15*time.Minute))
					return &img
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
			Reactions:    int(reactions),
			UserReacted:  hasUserReacted != 0,
			Comments:     comments,
			CommentCount: commentCount,
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

		// Fetch existing post to check current image
		query := sqlite.NewQuery(app.DB)
		existingPost, err := query.Posts.GetPostByID(r.Context(), postID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
			app.Logger.Error("failed to fetch post", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		var image sql.NullString

		if req.ImageID != nil {
			// ImageID field was provided
			if *req.ImageID != "" {
				// New image provided
				if existingPost.ImageID.Valid && *req.ImageID != existingPost.ImageID.String {
					// Old image exists and is different from new one - remove old and assign new
					err := app.File.RemoveImage(existingPost.ImageID.String)
					if err != nil {
						app.Logger.Error("failed to remove old image", "error", err.Error())
						utils.Internal(w, errors.New("failed to update something went wrong"))
						return
					}

					err = app.File.AssignImage(*req.ImageID)
					if err != nil {
						app.Logger.Error("failed to assign new image", "error", err.Error())
						utils.Internal(w, errors.New("failed to update something went wrong"))
						return
					}
				} else if !existingPost.ImageID.Valid {
					// No old image - just assign new one
					err = app.File.AssignImage(*req.ImageID)
					if err != nil {
						app.Logger.Error("failed to assign image", "error", err.Error())
						utils.Internal(w, errors.New("failed to update something went wrong"))
						return
					}
				}

				image = sql.NullString{Valid: true, String: *req.ImageID}
			} else {
				// Empty string provided - user wants to remove the image
				if existingPost.ImageID.Valid {
					err := app.File.RemoveImage(existingPost.ImageID.String)
					if err != nil {
						app.Logger.Error("failed to remove image", "error", err.Error())
						utils.Internal(w, errors.New("failed to update something went wrong"))
						return
					}
				}
				image = sql.NullString{Valid: false, String: ""}
			}
		} else {
			// ImageID field not provided - keep existing image
			if existingPost.ImageID.Valid {
				image = sql.NullString{Valid: true, String: existingPost.ImageID.String}
			} else {
				image = sql.NullString{Valid: false, String: ""}
			}
		}

		err = query.Posts.UpdatePost(r.Context(), db_posts.UpdatePostParams{
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

		post := helpers.FetchPostBasicInfo(app, postID, r.Context(), w)

		err = sqlite.NewQuery(app.DB).Posts.DeletePost(r.Context(), db_posts.DeletePostParams{
			PostID:   postID,
			AuthorID: currentUserID,
		})

		if err != nil {
			app.Logger.Error("failed to delete post", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// remove viewing permissions if post visibility was private
		if post.Visibility == "private" {
			err := sqlite.NewQuery(app.DB).Posts.RemoveViewingPermissionPostIDBatch(r.Context(), postID)
			if err != nil {
				app.Logger.Error("failed to remove viewing permissions", "error", err.Error())
				utils.Internal(w, errors.New("internal server error"))
			}
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

		// Fetch post basic info to check if viewing permissions need to be deleted
		post := helpers.FetchPostBasicInfo(app, postID, r.Context(), w)
		user := helpers.FetchUser(app, currentUserID, r.Context(), w)

		if post.Visibility == req.Visibility {
			utils.BadRequest(w, errors.New("visibility already set to "+req.Visibility+""))
			return
		}

		if user.IsPublic && req.Visibility == "semi-private" {
			utils.BadRequest(w, errors.New("cannot set semi-private visibility on public profile"))
			return
		} else if !user.IsPublic && req.Visibility == "public" {
			utils.BadRequest(w, errors.New("cannot set public visibility on private profile"))
			return
		}

		if post.Visibility == "private" {
			err := sqlite.NewQuery(app.DB).Posts.RemoveViewingPermissionPostIDBatch(r.Context(), postID)
			if err != nil {
				app.Logger.Error("failed to remove viewing permissions", "error", err.Error())
				utils.Internal(w, errors.New("internal server error"))
				return
			}
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

		if string(targetUserID) == string(post.AuthorID) {
			utils.BadRequest(w, errors.New("cannot add yourself to private post viewing list"))
			return
		}

		rowsAffected, err := sqlite.NewQuery(app.DB).Posts.AddPrivatePostViewingPermission(r.Context(), db_posts.AddPrivatePostViewingPermissionParams{
			UserID:     targetUserID,
			PostID:     postID,
			PostID_2:   postID,
			FollowerID: targetUserID,
		})

		if err != nil {
			if sqlite.CheckUniqueConstraint(err) {
				utils.BadRequest(w, errors.New("user already added to private post viewing list"))
				return
			}
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

		err = sqlite.NewQuery(app.DB).Posts.RemovePrivatePostViewingPermission(r.Context(), db_posts.RemovePrivatePostViewingPermissionParams{
			UserID: targetUserID,
			PostID: postID,
		})

		if err != nil {
			app.Logger.Error("failed to remove viewing permission", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, "User removed from private post viewing list")
	}
}

// -------------------- POSTS COMMENTS HANDLERS ---------------------

func GetComments(app *app.App) http.HandlerFunc {
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

		// Fetch post basic info and check visibility permissions
		postBasicInfo := helpers.FetchPostBasicInfo(app, postID, r.Context(), w)
		if postBasicInfo.PostID == nil {
			return
		}

		// Check if user can view this post's comments
		canView, err := helpers.CanViewPost(currentUserID, postBasicInfo.PostID, postBasicInfo.AuthorID, postBasicInfo.GroupID, postBasicInfo.Visibility, app, r)
		if err != nil {
			app.Logger.Error("failed to check post permissions", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if !canView {
			utils.Forbidden(w)
			return
		}

		// Get comments for the post
		comments, err := sqlite.NewQuery(app.DB).Posts.GetPostComments(r.Context(), db_posts.GetPostCommentsParams{
			PostID:   postID,
			AuthorID: currentUserID,
		})
		if err != nil {
			app.Logger.Error("failed to get comments", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Build response
		var commentsList []models.Comment
		for _, comment := range comments {
			commentUUID, _ := helpers.GenerateFromBytes(comment.CommentID)
			authorUUID, _ := helpers.GenerateFromBytes(comment.AuthorID)

			commentData := models.Comment{
				CommentID:       commentUUID,
				AuthorID:        authorUUID,
				AuthorFirstName: comment.FirstName,
				AuthorLastName:  comment.LastName,
				AuthorNickname: func() *string {
					if comment.Nickname.Valid {
						return &comment.Nickname.String
					}
					return nil
				}(),
				AuthorAvatar: func() *string {
					if comment.Avatar.Valid && comment.Avatar.String != "" {
						img := app.File.GenerateSignImage(comment.Avatar.String, currentUserID, time.Now().Add(15*time.Minute))
						return &img
					}
					return nil
				}(),
				Content: comment.Content,
				ImageID: &comment.ImageID.String,
				ImageUrl: func() string {
					if comment.ImageID.Valid && comment.FileName.Valid {
						path := app.File.GenerateSignImage(comment.FileName.String, currentUserID, time.Now().Add(15*time.Minute))
						return path
					}
					return ""
				}(),
				CreatedAt:   comment.CreatedAt.Time,
				Reactions:   int(comment.ReactionCount),
				UserReacted: comment.UserReacted != 0,
			}

			if comment.ParentCommentID != nil {
				parentUUID, err := helpers.GenerateFromBytes(comment.ParentCommentID)
				if err == nil {
					commentData.ParentCommentID = &parentUUID
				}
			}

			commentsList = append(commentsList, commentData)
		}

		utils.OK(w, commentsList)
	}
}

func CreateComment(app *app.App) http.HandlerFunc {
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

		var req models.CreateCommentRequest

		inputs := helpers.ValidateCreateComment.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		// Fetch post basic info and check visibility permissions
		postBasicInfo := helpers.FetchPostBasicInfo(app, postID, r.Context(), w)
		if postBasicInfo.PostID == nil {
			return
		}

		// Check if user can comment on this post (same rules as viewing)
		canComment, err := helpers.CanViewPost(currentUserID, postBasicInfo.PostID, postBasicInfo.AuthorID, postBasicInfo.GroupID, postBasicInfo.Visibility, app, r)
		if err != nil {
			app.Logger.Error("failed to check post permissions", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if !canComment {
			fmt.Println(canComment)
			utils.Forbidden(w)
			return
		}

		// Handle parent comment ID
		var parentCommentID []byte
		if req.ParentID != nil && *req.ParentID != "" {
			parentCommentID, err = helpers.GenerateFromString(*req.ParentID)
			if err != nil {
				utils.BadRequest(w, errors.New("invalid parent comment ID format"))
				return
			}

			parentComment, err := sqlite.NewQuery(app.DB).Posts.CheckCommentExists(r.Context(), parentCommentID)
			if err != nil {
				app.Logger.Error("failed to check parent comment", "error", err.Error())
				utils.Internal(w, errors.New("internal server error"))
				return
			}

			if parentComment == 0 {
				utils.BadRequest(w, errors.New("parent comment does not exist"))
				return
			}
		}

		// Generate new comment ID
		commentID, err := uuid.New().MarshalBinary()
		if err != nil {
			app.Logger.Error("failed to generate comment UUID", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Handle image ID
		var imageID sql.NullString
		if req.ImageID != nil && *req.ImageID != "" {
			imageID = sql.NullString{String: *req.ImageID, Valid: true}
			err = app.File.AssignImage(*req.ImageID)
			if err != nil {
				app.Logger.Error("failed to assign image", "error", err.Error())
				utils.Internal(w, errors.New("failed to assign image"))
				return
			}
		}

		// Create comment with visibility check
		rowsAffected, err := sqlite.NewQuery(app.DB).Posts.CreateComment(r.Context(), db_posts.CreateCommentParams{
			CommentID:       commentID,
			PostID:          postID,
			Content:         req.Content,
			ParentCommentID: parentCommentID,
			ImageID:         imageID,
			PostID_2:        postID,
			AuthorID:        currentUserID,
		})

		if err != nil {
			app.Logger.Error("failed to create comment", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if rowsAffected == 0 {
			fmt.Println("whatt?")
			utils.Forbidden(w)
			return
		}

		// Create notification for comment or reply
		if parentCommentID != nil {
			// This is a reply to a comment - notify the comment author
			parentCommentData, err := sqlite.NewQuery(app.DB).Posts.GetCommentById(r.Context(), parentCommentID)
			if err == nil && string(currentUserID) != string(parentCommentData.AuthorID) {
				err = helpers.CreateNotification(app, parentCommentData.AuthorID, constants.NotificationCommentReply, currentUserID, nil, nil, nil)
				if err != nil {
					app.Logger.Error("failed to create comment reply notification", "err", err)
					// Don't fail the request if notification fails
				}
			}
		} else if string(currentUserID) != string(postBasicInfo.AuthorID) {
			// This is a comment on a post - notify the post author (if not commenting on own post)
			err = helpers.CreateNotification(app, postBasicInfo.AuthorID, constants.NotificationPostComment, currentUserID, nil, nil, nil)
			if err != nil {
				app.Logger.Error("failed to create post comment notification", "err", err)
				// Don't fail the request if notification fails
			}
		}

		commentUUID, _ := helpers.GenerateFromBytes(commentID)
		authorID, _ := helpers.GenerateFromBytes(currentUserID)

		response := models.Comment{
			CommentID: commentUUID,
			AuthorID:  authorID,
			Content:   req.Content,
			CreatedAt: time.Now(),
			Reactions: 0,
		}

		if req.ParentID != nil && *req.ParentID != "" {
			response.ParentCommentID = req.ParentID
		}

		// Add image information to response if image was uploaded
		if imageID.Valid {
			response.ImageID = &imageID.String
			image, err := sqlite.NewQuery(app.DB).Image.GetImageById(r.Context(), imageID.String)
			if err == nil {
				response.ImageUrl = app.File.GenerateSignImage(image.FileName, currentUserID, time.Now().Add(15*time.Minute))
			}
		}

		utils.OK(w, response)
	}
}

func EditComment(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		commentIDHex := r.PathValue("commentId")
		if commentIDHex == "" {
			utils.BadRequest(w, errors.New("comment ID required"))
			return
		}

		commentID, err := helpers.GenerateFromString(commentIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid comment ID format"))
			return
		}

		var req models.UpdateCommentRequest

		inputs := helpers.ValidateUpdateComment.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		// Fetch existing comment to check current image
		query := sqlite.NewQuery(app.DB)
		existingComment, err := query.Posts.GetCommentById(r.Context(), commentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
			app.Logger.Error("failed to fetch comment", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Check ownership
		if string(existingComment.AuthorID) != string(currentUserID) {
			utils.Forbidden(w)
			return
		}

		var image sql.NullString

		if req.ImageID != nil {
			// ImageID field was provided
			if *req.ImageID != "" {
				// New image provided
				if existingComment.ImageID.Valid && *req.ImageID != existingComment.ImageID.String {
					// Old image exists and is different from new one - remove old and assign new
					err := app.File.RemoveImage(existingComment.ImageID.String)
					if err != nil {
						app.Logger.Error("failed to remove old image", "error", err.Error())
						utils.Internal(w, errors.New("failed to update something went wrong"))
						return
					}

					err = app.File.AssignImage(*req.ImageID)
					if err != nil {
						app.Logger.Error("failed to assign new image", "error", err.Error())
						utils.Internal(w, errors.New("failed to update something went wrong"))
						return
					}
				} else if !existingComment.ImageID.Valid {
					// No old image - just assign new one
					err = app.File.AssignImage(*req.ImageID)
					if err != nil {
						app.Logger.Error("failed to assign image", "error", err.Error())
						utils.Internal(w, errors.New("failed to update something went wrong"))
						return
					}
				}

				image = sql.NullString{Valid: true, String: *req.ImageID}
			} else {
				// Empty string provided - user wants to remove the image
				if existingComment.ImageID.Valid {
					err := app.File.RemoveImage(existingComment.ImageID.String)
					if err != nil {
						app.Logger.Error("failed to remove image", "error", err.Error())
						utils.Internal(w, errors.New("failed to update something went wrong"))
						return
					}
				}
				image = sql.NullString{Valid: false, String: ""}
			}
		} else {
			// ImageID field not provided - keep existing image
			if existingComment.ImageID.Valid {
				image = sql.NullString{Valid: true, String: existingComment.ImageID.String}
			} else {
				image = sql.NullString{Valid: false, String: ""}
			}
		}

		// Update comment (only owner can edit)
		err = query.Posts.EditComment(r.Context(), db_posts.EditCommentParams{
			Content:   req.Content,
			ImageID:   image,
			CommentID: commentID,
			AuthorID:  currentUserID,
		})

		if err != nil {
			app.Logger.Error("failed to edit comment", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]string{
			"message":    "Comment updated successfully",
			"comment_id": commentIDHex,
		})
	}
}

func DeleteComment(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		commentIDHex := r.PathValue("commentId")
		if commentIDHex == "" {
			utils.BadRequest(w, errors.New("comment ID required"))
			return
		}

		commentID, err := helpers.GenerateFromString(commentIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid comment ID format"))
			return
		}

		hasChildren, err := sqlite.NewQuery(app.DB).Posts.CheckCommentHasChildren(r.Context(), commentID)
		if err != nil {
			app.Logger.Error("failed to check comment has_children", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if hasChildren != 0 {
			utils.BadRequest(w, errors.New("comment has replies"))
			return
		}

		// Delete comment (only owner can delete)
		err = sqlite.NewQuery(app.DB).Posts.DeleteComment(r.Context(), db_posts.DeleteCommentParams{
			CommentID: commentID,
			AuthorID:  currentUserID,
		})

		if err != nil {
			app.Logger.Error("failed to delete comment", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]string{
			"message":    "Comment deleted successfully",
			"comment_id": commentIDHex,
		})
	}
}

// -------------------- POSTS REACTIONS HANDLERS ---------------------

func GetReactions(app *app.App) http.HandlerFunc {
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

		// Fetch post basic info and check visibility permissions
		postBasicInfo := helpers.FetchPostBasicInfo(app, postID, r.Context(), w)
		if postBasicInfo.PostID == nil {
			return
		}

		// Check if user can view this post's reactions
		canView, err := helpers.CanViewPost(currentUserID, postBasicInfo.PostID, postBasicInfo.AuthorID, postBasicInfo.GroupID, postBasicInfo.Visibility, app, r)
		if err != nil {
			app.Logger.Error("failed to check post permissions", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if !canView {
			utils.Forbidden(w)
			return
		}

		// Get reaction count for the post
		count, err := sqlite.NewQuery(app.DB).Posts.GetPostReactions(r.Context(), postID)
		if err != nil {
			app.Logger.Error("failed to get post reactions", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Check if current user has reacted
		hasReacted, err := sqlite.NewQuery(app.DB).Posts.HasUserReacted(r.Context(), db_posts.HasUserReactedParams{
			AuthorID:   currentUserID,
			TargetType: "post",
			TargetID:   postID,
		})
		if err != nil {
			app.Logger.Error("failed to check user reaction", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]interface{}{
			"count":        count,
			"user_reacted": hasReacted > 0,
		})
	}
}

func ToggleReaction(app *app.App) http.HandlerFunc {
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

		// Fetch post basic info and check visibility permissions
		postBasicInfo := helpers.FetchPostBasicInfo(app, postID, r.Context(), w)
		if postBasicInfo.PostID == nil {
			return
		}

		// Check if user can react to this post (same rules as viewing)
		canReact, err := helpers.CanViewPost(currentUserID, postBasicInfo.PostID, postBasicInfo.AuthorID, postBasicInfo.GroupID, postBasicInfo.Visibility, app, r)
		if err != nil {
			app.Logger.Error("failed to check post permissions", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}
		if !canReact {
			utils.Forbidden(w)
			return
		}

		commentIDHex := r.PathValue("commentId")

		targetType := "post"
		targetID := postID

		if commentIDHex != "" {
			targetType = "comment"
			targetID, err = helpers.GenerateFromString(commentIDHex)
			if err != nil {
				utils.BadRequest(w, errors.New("invalid comment ID format"))
				return
			}
		}

		// Check if user has already reacted
		hasReacted, err := sqlite.NewQuery(app.DB).Posts.HasUserReacted(r.Context(), db_posts.HasUserReactedParams{
			AuthorID:   currentUserID,
			TargetType: targetType,
			TargetID:   targetID,
		})
		if err != nil {
			app.Logger.Error("failed to check user reaction", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if hasReacted > 0 {
			// User has reacted, so remove the reaction
			err = sqlite.NewQuery(app.DB).Posts.DeleteReaction(r.Context(), db_posts.DeleteReactionParams{
				AuthorID:   currentUserID,
				TargetType: targetType,
				TargetID:   targetID,
			})
			if err != nil {
				app.Logger.Error("failed to delete reaction", "error", err.Error())
				utils.Internal(w, errors.New("internal server error"))
				return
			}

			utils.OK(w, map[string]interface{}{
				"message":      "Reaction removed",
				"user_reacted": false,
			})
		} else {
			// User has not reacted, so add a reaction
			reactionID, err := uuid.New().MarshalBinary()
			if err != nil {
				app.Logger.Error("failed to generate reaction UUID", "error", err.Error())
				utils.Internal(w, errors.New("internal server error"))
				return
			}

			rowsAffected, err := sqlite.NewQuery(app.DB).Posts.CreateReaction(r.Context(), db_posts.CreateReactionParams{
				ReactionID: reactionID,
				TargetType: targetType,
				TargetID:   targetID,
				AuthorID:   currentUserID,
				PostID:     postID,
			})

			if err != nil {
				app.Logger.Error("failed to create reaction", "error", err.Error())
				utils.Internal(w, errors.New("internal server error"))
				return
			}

			if rowsAffected == 0 {
				utils.Forbidden(w)
				return
			}

			// Create notification for reaction
			var notifType constants.NotificationType
			var notifReceiver []byte

			if targetType == "post" {
				notifType = constants.NotificationPostReaction
				notifReceiver = postBasicInfo.AuthorID
			} else if targetType == "comment" {
				notifType = constants.NotificationCommentReaction
				commentData, err := sqlite.NewQuery(app.DB).Posts.GetCommentById(r.Context(), targetID)
				if err != nil {
					app.Logger.Error("failed to get comment for notification", "err", err)
				} else {
					notifReceiver = commentData.AuthorID
				}
			}

			// Don't notify if reacting to own content
			if len(notifReceiver) > 0 && string(currentUserID) != string(notifReceiver) {
				err = helpers.CreateNotification(app, notifReceiver, notifType, currentUserID, nil, nil, nil)
				if err != nil {
					app.Logger.Error("failed to create reaction notification", "err", err, "type", notifType)
					// Don't fail the request if notification fails
				}
			}

			utils.OK(w, map[string]interface{}{
				"message":      "Reaction added",
				"user_reacted": true,
			})
		}
	}
}
