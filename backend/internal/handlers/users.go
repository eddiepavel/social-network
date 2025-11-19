package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"social-network/internal/auth"
	"social-network/internal/models"
	"strings"
	"time"
)

// UsersHandler handles user-related requests
type UsersHandler struct {
	db *sql.DB
}

// NewUsersHandler creates a new UsersHandler
func NewUsersHandler(db *sql.DB) *UsersHandler {
	return &UsersHandler{db: db}
}

// GetUserProfile handles GET /api/users/:id
// Returns user profile respecting privacy settings
func (h *UsersHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	currentUserID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract user ID from URL path: /api/users/:id
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/users/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	targetUserIDHex := pathParts[0]
	targetUserID, err := hex.DecodeString(targetUserIDHex)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Check if user can view this profile
	canView, err := h.canViewProfile(currentUserID, targetUserID)
	if err != nil {
		log.Printf("Error checking profile access: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !canView {
		http.Error(w, "You don't have permission to view this profile", http.StatusForbidden)
		return
	}

	// Fetch user profile
	user, err := h.getUserByID(targetUserID)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return user response
	response := h.userToResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProfile handles PUT /api/users/profile
// Updates current user's profile (own profile only)
func (h *UsersHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}

	if req.FirstName != nil {
		updates = append(updates, "first_name = ?")
		args = append(args, *req.FirstName)
	}
	if req.LastName != nil {
		updates = append(updates, "last_name = ?")
		args = append(args, *req.LastName)
	}
	if req.Nickname != nil {
		updates = append(updates, "nickname = ?")
		args = append(args, *req.Nickname)
	}
	if req.AboutMe != nil {
		updates = append(updates, "about_me = ?")
		args = append(args, *req.AboutMe)
	}
	if req.Avatar != nil {
		updates = append(updates, "avatar = ?")
		args = append(args, *req.Avatar)
	}

	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// Add user_id to args
	args = append(args, userID)

	// Execute update
	query := "UPDATE users SET " + strings.Join(updates, ", ") + " WHERE user_id = ?"
	_, err := h.db.Exec(query, args...)
	if err != nil {
		log.Printf("Failed to update user profile: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// Fetch updated user
	user, err := h.getUserByID(userID)
	if err != nil {
		log.Printf("Failed to fetch updated user: %v", err)
		http.Error(w, "Profile updated but failed to fetch details", http.StatusInternalServerError)
		return
	}

	// Return updated user
	response := h.userToResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdatePrivacy handles PUT /api/users/privacy
// Updates current user's privacy settings (own profile only)
func (h *UsersHandler) UpdatePrivacy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.UpdatePrivacyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update privacy setting
	query := "UPDATE users SET is_public = ? WHERE user_id = ?"
	_, err := h.db.Exec(query, req.IsPublic, userID)
	if err != nil {
		log.Printf("Failed to update privacy setting: %v", err)
		http.Error(w, "Failed to update privacy", http.StatusInternalServerError)
		return
	}

	// Fetch updated user
	user, err := h.getUserByID(userID)
	if err != nil {
		log.Printf("Failed to fetch updated user: %v", err)
		http.Error(w, "Privacy updated but failed to fetch details", http.StatusInternalServerError)
		return
	}

	// Return updated user
	response := h.userToResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

// canViewProfile checks if currentUser can view targetUser's profile
// Rules:
// - Can always view own profile
// - Can view public profiles
// - Can view private profiles if following them (accepted follower)
func (h *UsersHandler) canViewProfile(currentUserID, targetUserID []byte) (bool, error) {
	// Can always view own profile
	if string(currentUserID) == string(targetUserID) {
		return true, nil
	}

	// Check if target user's profile is public
	var isPublic bool
	query := "SELECT is_public FROM users WHERE user_id = ?"
	err := h.db.QueryRow(query, targetUserID).Scan(&isPublic)
	if err != nil {
		return false, err
	}

	// Public profiles are visible to everyone
	if isPublic {
		return true, nil
	}

	// For private profiles, check if current user is a follower
	query = `
		SELECT COUNT(*) FROM followers
		WHERE follower_id = ? AND followee_id = ? AND status = 'accepted'
	`
	var count int
	err = h.db.QueryRow(query, currentUserID, targetUserID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (h *UsersHandler) getUserByID(userID []byte) (*models.User, error) {
	query := `
		SELECT user_id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public, created_at
		FROM users
		WHERE user_id = ?
	`

	user := &models.User{}
	err := h.db.QueryRow(query, userID).Scan(
		&user.UserID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.DOB,
		&user.Avatar,
		&user.Nickname,
		&user.AboutMe,
		&user.IsPublic,
		&user.CreatedAt,
	)

	return user, err
}

func (h *UsersHandler) userToResponse(user *models.User) models.UserResponse {
	return models.UserResponse{
		UserID:    hex.EncodeToString(user.UserID),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		DOB:       user.DOB,
		Avatar:    user.Avatar,
		Nickname:  user.Nickname,
		AboutMe:   user.AboutMe,
		IsPublic:  user.IsPublic,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}
