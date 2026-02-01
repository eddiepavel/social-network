package models

import "time"

// FollowerResponse represents a follower/following user
type FollowerResponse struct {
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Avatar    *string   `json:"avatar,omitempty"`
	Nickname  *string   `json:"nickname,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type FollowRequestsResponse struct {
	ID           int64     `json:"id"`
	FollowerID   string    `json:"follower_id"`
	FollowerName string    `json:"follower_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// FollowStatusResponse represents the follow status between two users
type FollowStatusResponse struct {
	Status string `json:"status"` // "following", "requested", "none", "self"
}
