package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
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
		for _, comment := range postComments {
			commentAuthorID, _ := helpers.GenerateFromBytes(comment.AuthorID)
			commentID, _ := helpers.GenerateFromBytes(comment.CommentID)
			var parentID string
			if comment.ParentCommentID != nil {
				parentID, _ = helpers.GenerateFromBytes(comment.ParentCommentID)
			}
			comments = append(comments, models.Comment{
				CommentID:       commentID,
				AuthorID:        commentAuthorID,
				Content:         comment.Content,
				ParentCommentID: &parentID,
				ImageID:         &comment.ImageID.String,
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
			PostID:     postIDHex,
			AuthorID:   authorUuid,
			Content:    post.Content,
			ImageID:    &post.ImageID.String,
			Visibility: post.Visibility,
			CreatedAt: func() time.Time {
				if post.CreatedAt.Valid {
					return post.CreatedAt.Time
				}
				return time.Time{}
			}(),
			Reactions:   int(reactions),
			UserReacted: hasUserReacted != 0,
			Comments:    comments,
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
				CommentID:   commentUUID,
				AuthorID:    authorUUID,
				Content:     comment.Content,
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

			// TODO: Add image support
			// if comment.ImageID.Valid {
			// 	commentData.ImageID = &comment.ImageID.String
			// }

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
		canComment, err := helpers.CanViewPost(currentUserID, postBasicInfo.PostID, postBasicInfo.AuthorID, postBasicInfo.Visibility, app, r)
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

		// TODO: Handle image ID when image service is ready
		var imageID sql.NullString
		// if req.ImageID != nil && *req.ImageID != "" {
		// 	imageID = sql.NullString{String: *req.ImageID, Valid: true}
		// }

		// Create comment with visibility check
		rowsAffected, err := sqlite.NewQuery(app.DB).Posts.CreateComment(r.Context(), db_posts.CreateCommentParams{
			CommentID:       commentID,
			PostID:          postID,
			UserID:          currentUserID,
			Content:         req.Content,
			ParentCommentID: parentCommentID,
			ImageID:         imageID,
			PostID_2:        postID,
			FollowerID:      currentUserID,
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

		// Update comment (only owner can edit)
		err = sqlite.NewQuery(app.DB).Posts.EditComment(r.Context(), db_posts.EditCommentParams{
			Content:   req.Content,
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
		canReact, err := helpers.CanViewPost(currentUserID, postBasicInfo.PostID, postBasicInfo.AuthorID, postBasicInfo.Visibility, app, r)
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
				FollowerID: currentUserID,
				UserID:     currentUserID,
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

			utils.OK(w, map[string]interface{}{
				"message":      "Reaction added",
				"user_reacted": true,
			})
		}
	}
}
