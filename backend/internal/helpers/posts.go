package helpers

import (
	"context"
	"database/sql"
	"net/http"
	"social-network/app"
	"social-network/internal/utils"
	db_followers "social-network/pkg/db/queries/followers"
	db_posts "social-network/pkg/db/queries/posts"
	"social-network/pkg/db/sqlite"
)

// FetchPostBasicInfo fetches basic post info (post_id, author_id, visibility) for permission checks
func FetchPostBasicInfo(app *app.App, postID []byte, ctx context.Context, w http.ResponseWriter) db_posts.GetPostBasicInfoRow {
	post, err := sqlite.NewQuery(app.DB).Posts.GetPostBasicInfo(ctx, postID)
	if err == sql.ErrNoRows {
		utils.NotFound(w)
		return db_posts.GetPostBasicInfoRow{}
	}
	if err != nil {
		app.Logger.Error("Failed to fetch post", "error", err.Error())
		utils.Internal(w, err)
		return db_posts.GetPostBasicInfoRow{}
	}

	return post
}

// FetchPost fetches a complete post with reactions and comments
func FetchPost(app *app.App, postID []byte, ctx context.Context, w http.ResponseWriter) db_posts.Post {
	post, err := sqlite.NewQuery(app.DB).Posts.GetPostByID(ctx, postID)
	if err == sql.ErrNoRows {
		utils.NotFound(w)
		return db_posts.Post{}
	}
	if err != nil {
		app.Logger.Error("Failed to fetch post", "error", err.Error())
		utils.Internal(w, err)
		return db_posts.Post{}
	}

	return post
}

// CanViewPost checks if currentUser can view the post
// Rules:
// - Can view public posts
// - Can view private posts if following the author
// - Can view private posts if explicit viewing permission granted
func CanViewPost(currentUserID []byte, postID []byte, authorID []byte, visibility string, app *app.App, r *http.Request) (bool, error) {
	// Can always view own posts
	if string(currentUserID) == string(authorID) {
		return true, nil
	}

	// Public posts are visible to everyone
	if visibility == "public" {
		return true, nil
	}

	// Check if current user is following the post author
	_, err := sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(), db_followers.CheckIfUserFollowsParams{
		FollowerID: currentUserID,
		FolloweeID: authorID,
	})

	// If following, can view
	if err == nil && visibility == "semi-private" {
		return true, nil
	}

	// If not following and error is not "not found", return the error
	if err != sql.ErrNoRows && err != nil {
		return false, err
	}

	// Check for explicit viewing permission
	_, err = sqlite.NewQuery(app.DB).Posts.CheckPrivatePostUserPermit(r.Context(), db_posts.CheckPrivatePostUserPermitParams{
		UserID: currentUserID,
		PostID: postID,
	})

	if err == nil {
		return true, nil
	}

	if err != sql.ErrNoRows {
		return false, err
	}

	return false, nil
}
