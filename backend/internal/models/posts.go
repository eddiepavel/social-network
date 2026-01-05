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
	Content    string `json:"content"`
	ImageID    string `json:"image_id"`
	Visibility string `json:"visibility"`
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
	Content string `json:"content"`
	ImageID string `json:"image_id"`
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
	PostID        string    `json:"post_id"`
	Content       string    `json:"content"`
	ImageID       string    `json:"image_id"`
	ImageUrl      string    `json:"image_url"`
	Visibility    string    `json:"visibility"`
	AuthorID      string    `json:"author_id"`
	CreatedAt     time.Time `json:"created_at"`
	ReactionCount int64     `json:"reaction_count"`
	CommentCount  int64     `json:"comment_count"`
}

type Reaction struct {
	ReactionID   string `json:"reaction_id"`
	UserID       []byte `json:"user_id"`
	ReactionType string `json:"reaction_type"`
}

type Comment struct {
	CommentID       string     `json:"comment_id"`
	UserID          []byte     `json:"user_id"`
	Content         string     `json:"content"`
	ParentCommentID *string    `json:"parent_comment_id"`
	ImageID         *string    `json:"image_id"`
	CreatedAt       time.Time  `json:"created_at"`
	Reactions       []Reaction `json:"reactions"`
}

type PostWithCommentsReactionsResponse struct {
	PostID     string     `json:"post_id"`
	Content    string     `json:"content"`
	ImageID    string     `json:"image_id"`
	Visibility string     `json:"visibility"`
	AuthorID   []byte     `json:"author_id"`
	CreatedAt  time.Time  `json:"created_at"`
	Reactions  []Reaction `json:"reactions"`
	Comments   []Comment  `json:"comments"`
}
