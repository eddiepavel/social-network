package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"social-network/internal/auth"
	"strings"
	"time"
)

// FollowersHandler handles follower-related requests
type FollowersHandler struct {
	db *sql.DB
}

// NewFollowersHandler creates a new FollowersHandler
func NewFollowersHandler(db *sql.DB) *FollowersHandler {
	return &FollowersHandler{db: db}
}

// FollowerResponse represents a follower/following user
type FollowerResponse struct {
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Avatar    *string   `json:"avatar"`
	Nickname  *string   `json:"nickname"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// FollowUser handles POST /api/follow/:userId
// Sends follow request or auto-follows if public profile
func (h *FollowersHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	currentUserID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract target user ID from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/follow/"), "/")
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

	// Can't follow yourself
	if string(currentUserID) == string(targetUserID) {
		http.Error(w, "Cannot follow yourself", http.StatusBadRequest)
		return
	}

	// Check if target user exists and is public
	var isPublic bool
	err = h.db.QueryRow("SELECT is_public FROM users WHERE user_id = ?", targetUserID).Scan(&isPublic)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if follow relationship already exists
	var existingStatus string
	err = h.db.QueryRow(
		"SELECT status FROM followers WHERE follower_id = ? AND followee_id = ?",
		currentUserID, targetUserID,
	).Scan(&existingStatus)

	if err == nil {
		// Relationship exists
		if existingStatus == "accepted" {
			http.Error(w, "Already following this user", http.StatusConflict)
			return
		}
		if existingStatus == "pending" {
			http.Error(w, "Follow request already pending", http.StatusConflict)
			return
		}
		// If rejected, we can update it
	}

	// Determine status based on profile visibility
	status := "pending"
	if isPublic {
		status = "accepted"
	}

	// Insert or update follow relationship
	if err == sql.ErrNoRows {
		// Insert new relationship
		_, err = h.db.Exec(
			"INSERT INTO followers (follower_id, followee_id, status) VALUES (?, ?, ?)",
			currentUserID, targetUserID, status,
		)
	} else {
		// Update existing relationship
		_, err = h.db.Exec(
			"UPDATE followers SET status = ?, created_at = CURRENT_TIMESTAMP WHERE follower_id = ? AND followee_id = ?",
			status, currentUserID, targetUserID,
		)
	}

	if err != nil {
		log.Printf("Failed to create follow relationship: %v", err)
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"follower_id": hex.EncodeToString(currentUserID),
		"followee_id": targetUserIDHex,
		"status":      status,
		"message":     getFollowMessage(status),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UnfollowUser handles DELETE /api/follow/:userId
func (h *FollowersHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	currentUserID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract target user ID from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/follow/"), "/")
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

	// Delete follow relationship
	result, err := h.db.Exec(
		"DELETE FROM followers WHERE follower_id = ? AND followee_id = ?",
		currentUserID, targetUserID,
	)
	if err != nil {
		log.Printf("Failed to unfollow user: %v", err)
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Not following this user", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Unfollowed successfully"))
}

// AcceptFollowRequest handles POST /api/follow/accept/:followerId
func (h *FollowersHandler) AcceptFollowRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context (the followee)
	currentUserID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract follower ID from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/follow/accept/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Follower ID required", http.StatusBadRequest)
		return
	}

	followerIDHex := pathParts[0]
	followerID, err := hex.DecodeString(followerIDHex)
	if err != nil {
		http.Error(w, "Invalid follower ID format", http.StatusBadRequest)
		return
	}

	// Update follow request status to accepted
	result, err := h.db.Exec(
		"UPDATE followers SET status = 'accepted' WHERE follower_id = ? AND followee_id = ? AND status = 'pending'",
		followerID, currentUserID,
	)
	if err != nil {
		log.Printf("Failed to accept follow request: %v", err)
		http.Error(w, "Failed to accept follow request", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No pending follow request found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Follow request accepted"))
}

// RejectFollowRequest handles POST /api/follow/reject/:followerId
func (h *FollowersHandler) RejectFollowRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context (the followee)
	currentUserID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract follower ID from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/follow/reject/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Follower ID required", http.StatusBadRequest)
		return
	}

	followerIDHex := pathParts[0]
	followerID, err := hex.DecodeString(followerIDHex)
	if err != nil {
		http.Error(w, "Invalid follower ID format", http.StatusBadRequest)
		return
	}

	// Update follow request status to rejected
	result, err := h.db.Exec(
		"UPDATE followers SET status = 'rejected' WHERE follower_id = ? AND followee_id = ? AND status = 'pending'",
		followerID, currentUserID,
	)
	if err != nil {
		log.Printf("Failed to reject follow request: %v", err)
		http.Error(w, "Failed to reject follow request", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No pending follow request found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Follow request rejected"))
}

// GetFollowers handles GET /api/followers/:userId
func (h *FollowersHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
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

	// Extract target user ID from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/followers/"), "/")
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

	// Check if current user can view this user's followers
	canView, err := h.canViewFollowers(currentUserID, targetUserID)
	if err != nil {
		log.Printf("Error checking followers access: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !canView {
		http.Error(w, "You don't have permission to view these followers", http.StatusForbidden)
		return
	}

	// Get followers list
	query := `
		SELECT u.user_id, u.first_name, u.last_name, u.avatar, u.nickname, f.status, f.created_at
		FROM followers f
		JOIN users u ON f.follower_id = u.user_id
		WHERE f.followee_id = ? AND f.status = 'accepted'
		ORDER BY f.created_at DESC
	`

	rows, err := h.db.Query(query, targetUserID)
	if err != nil {
		log.Printf("Failed to fetch followers: %v", err)
		http.Error(w, "Failed to fetch followers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	followers := []FollowerResponse{}
	for rows.Next() {
		var f FollowerResponse
		var userID []byte
		err := rows.Scan(&userID, &f.FirstName, &f.LastName, &f.Avatar, &f.Nickname, &f.Status, &f.CreatedAt)
		if err != nil {
			log.Printf("Failed to scan follower: %v", err)
			continue
		}
		f.UserID = hex.EncodeToString(userID)
		followers = append(followers, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

// GetFollowing handles GET /api/following/:userId
func (h *FollowersHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
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

	// Extract target user ID from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/following/"), "/")
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

	// Check if current user can view this user's following list
	canView, err := h.canViewFollowers(currentUserID, targetUserID)
	if err != nil {
		log.Printf("Error checking following access: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !canView {
		http.Error(w, "You don't have permission to view this following list", http.StatusForbidden)
		return
	}

	// Get following list
	query := `
		SELECT u.user_id, u.first_name, u.last_name, u.avatar, u.nickname, f.status, f.created_at
		FROM followers f
		JOIN users u ON f.followee_id = u.user_id
		WHERE f.follower_id = ? AND f.status = 'accepted'
		ORDER BY f.created_at DESC
	`

	rows, err := h.db.Query(query, targetUserID)
	if err != nil {
		log.Printf("Failed to fetch following: %v", err)
		http.Error(w, "Failed to fetch following", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	following := []FollowerResponse{}
	for rows.Next() {
		var f FollowerResponse
		var userID []byte
		err := rows.Scan(&userID, &f.FirstName, &f.LastName, &f.Avatar, &f.Nickname, &f.Status, &f.CreatedAt)
		if err != nil {
			log.Printf("Failed to scan following: %v", err)
			continue
		}
		f.UserID = hex.EncodeToString(userID)
		following = append(following, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(following)
}

// GetFollowRequests handles GET /api/follow/requests
// Returns pending follow requests for the current user
func (h *FollowersHandler) GetFollowRequests(w http.ResponseWriter, r *http.Request) {
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

	// Get pending follow requests
	query := `
		SELECT u.user_id, u.first_name, u.last_name, u.avatar, u.nickname, f.status, f.created_at
		FROM followers f
		JOIN users u ON f.follower_id = u.user_id
		WHERE f.followee_id = ? AND f.status = 'pending'
		ORDER BY f.created_at DESC
	`

	rows, err := h.db.Query(query, currentUserID)
	if err != nil {
		log.Printf("Failed to fetch follow requests: %v", err)
		http.Error(w, "Failed to fetch follow requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	requests := []FollowerResponse{}
	for rows.Next() {
		var f FollowerResponse
		var userID []byte
		err := rows.Scan(&userID, &f.FirstName, &f.LastName, &f.Avatar, &f.Nickname, &f.Status, &f.CreatedAt)
		if err != nil {
			log.Printf("Failed to scan follow request: %v", err)
			continue
		}
		f.UserID = hex.EncodeToString(userID)
		requests = append(requests, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// Helper functions

func (h *FollowersHandler) canViewFollowers(currentUserID, targetUserID []byte) (bool, error) {
	// Can always view own followers/following
	if string(currentUserID) == string(targetUserID) {
		return true, nil
	}

	// Check if target user's profile is public
	var isPublic bool
	err := h.db.QueryRow("SELECT is_public FROM users WHERE user_id = ?", targetUserID).Scan(&isPublic)
	if err != nil {
		return false, err
	}

	if isPublic {
		return true, nil
	}

	// For private profiles, check if current user is a follower
	var count int
	err = h.db.QueryRow(
		"SELECT COUNT(*) FROM followers WHERE follower_id = ? AND followee_id = ? AND status = 'accepted'",
		currentUserID, targetUserID,
	).Scan(&count)

	return count > 0, err
}

func getFollowMessage(status string) string {
	if status == "accepted" {
		return "Now following user"
	}
	return "Follow request sent"
}
