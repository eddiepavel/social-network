package models

import "time"

type Post struct {
	PostID     []byte    `json:"post_id"`
	Content    string    `json:"content"`
	ImageID    string    `json:"image_id"`
	Visibility string    `json:"visibility"`
	AuthorID   []byte    `json:"author_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreatePostRequest struct {
	Content      string   `json:"content"`
	ImageID      string   `json:"image_id"`
	Visibility   string   `json:"visibility"`
	AllowedUsers []string `json:"allowed_users"` // User IDs who can view private posts
}

type EditPostRequest struct {
	PostID     []byte    `json:"post_id"`
	Content    string    `json:"content"`
	ImageID    string    `json:"image_id"`
	Visibility string    `json:"visibility"`
	CreatedAt  time.Time `json:"created_at"`
}

type PostResponse struct {
	PostID     string    `json:"post_id"`
	Content    string    `json:"content"`
	ImageID    string    `json:"image_id"`
	Visibility string    `json:"visibility"`
	AuthorID   []byte    `json:"author_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type UpdatePostRequest struct {
	Content string  `json:"content"`
	ImageID *string `json:"image_id"`
}

type UpdateVisibilityRequest struct {
	Visibility string `json:"visibility"`
}

type AddUserToPrivatePostRequest struct {
	UserID string `json:"user_id"`
}

type RemoveUserFromPrivatePostRequest struct {
	UserID string `json:"user_id"`
}

type FeedPostResponse struct {
	PostID          string    `json:"post_id"`
	Content         string    `json:"content"`
	ImageID         *string   `json:"image_id"`
	ImageUrl        string    `json:"image_url"`
	Visibility      string    `json:"visibility"`
	AuthorID        string    `json:"author_id"`
	AuthorFirstName string    `json:"author_first_name,omitempty"`
	AuthorLastName  string    `json:"author_last_name,omitempty"`
	AuthorNickname  *string   `json:"author_nickname,omitempty"`
	AuthorAvatar    *string   `json:"author_avatar,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	ReactionCount   int64     `json:"reaction_count"`
	UserReacted     bool      `json:"user_reacted"`
	CommentCount    int64     `json:"comment_count"`
}

type Comment struct {
	CommentID       string    `json:"comment_id"`
	AuthorID        string    `json:"author_id"`
	AuthorFirstName string    `json:"author_first_name,omitempty"`
	AuthorLastName  string    `json:"author_last_name,omitempty"`
	AuthorNickname  *string   `json:"author_nickname,omitempty"`
	AuthorAvatar    *string   `json:"author_avatar,omitempty"`
	Content         string    `json:"content"`
	ParentCommentID *string   `json:"parent_comment_id"`
	ImageID         *string   `json:"image_id"`
	ImageUrl        string    `json:"image_url"`
	CreatedAt       time.Time `json:"created_at"`
	Reactions       int       `json:"reactions"`
	UserReacted     bool      `json:"user_reacted"`
}

type PostWithCommentsReactionsResponse struct {
	PostID          string    `json:"post_id"`
	Content         string    `json:"content"`
	ImageID         *string   `json:"image_id"`
	ImageUrl        string    `json:"image_url"`
	Visibility      string    `json:"visibility"`
	AuthorID        string    `json:"author_id"`
	AuthorFirstName string    `json:"author_first_name,omitempty"`
	AuthorLastName  string    `json:"author_last_name,omitempty"`
	AuthorNickname  *string   `json:"author_nickname,omitempty"`
	AuthorAvatar    *string   `json:"author_avatar,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	Reactions       int       `json:"reactions"`
	UserReacted     bool      `json:"user_reacted"`
	Comments        []Comment `json:"comments"`
}

type CreateCommentRequest struct {
	Content  string  `json:"content"`
	ParentID *string `json:"parent_id"`
	// TODO: Add image support when image service is ready
	ImageID *string `json:"image_id"`
}

type UpdateCommentRequest struct {
	Content string  `json:"content"`
	ImageID *string `json:"image_id"`
}

type DeleteCommentRequest struct {
}
