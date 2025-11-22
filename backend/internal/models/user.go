package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	UserID       []byte    `json:"user_id"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"` // Never send password hash to client
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	DOB          string    `json:"dob"`
	Avatar       string    `json:"avatar"`
	Nickname     string    `json:"nickname"`
	AboutMe      string    `json:"about_me"`
	IsPublic     bool      `json:"is_public"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateUserRequest represents the request body for user registration
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	DOB       string `json:"dob"`
	Avatar    string `json:"avatar"`
	Nickname  string `json:"nickname"`
	AboutMe   string `json:"about_me"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse represents the user data returned to the client
// better to omit empty if we have null value
type UserResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	DOB       string `json:"dob"`
	Avatar    string `json:"avatar"`
	Nickname  string `json:"nickname"`
	AboutMe   string `json:"about_me"`
	IsPublic  bool   `json:"is_public"`
	CreatedAt string `json:"created_at"`
}

// UpdateProfileRequest represents the request body for updating user profile
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Nickname  *string `json:"nickname"`
	AboutMe   *string `json:"about_me"`
	Avatar    *string `json:"avatar"`
}

// UpdatePrivacyRequest represents the request body for updating privacy settings
type UpdatePrivacyRequest struct {
	IsPublic bool `json:"is_public"`
}
