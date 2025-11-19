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

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db *sql.DB
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.DOB == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Generate UUID for user
	userID, err := uuid.New().MarshalBinary()
	if err != nil {
		log.Printf("Failed to generate UUID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert user into database
	query := `
		INSERT INTO users (user_id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`

	_, err = h.db.Exec(query, userID, req.Email, hashedPassword, req.FirstName, req.LastName, req.DOB, req.Avatar, req.Nickname, req.AboutMe)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		log.Printf("Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Fetch the created user
	user, err := h.getUserByID(userID)
	if err != nil {
		log.Printf("Failed to fetch created user: %v", err)
		http.Error(w, "User created but failed to fetch details", http.StatusInternalServerError)
		return
	}

	// Create session for auto-login
	sessionID, err := auth.CreateSession(h.db, userID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "User created but failed to create session", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	auth.SetSessionCookie(w, sessionID)

	// Return user response
	response := h.userToResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Missing email or password", http.StatusBadRequest)
		return
	}

	// Fetch user by email
	query := `SELECT user_id, email, password_hash FROM users WHERE email = ?`
	var userID []byte
	var email string
	var passwordHash []byte

	err := h.db.QueryRow(query, req.Email).Scan(&userID, &email, &passwordHash)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Compare password with bcrypt
	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Fetch full user details
	user, err := h.getUserByID(userID)
	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID, err := auth.CreateSession(h.db, userID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	auth.SetSessionCookie(w, sessionID)

	// Return user response
	response := h.userToResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSession returns the current user from session cookie (protected endpoint)
func (h *AuthHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user_id from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch user
	user, err := h.getUserByID(userID)
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

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session cookie
	sessionID, err := auth.GetSessionCookie(r)
	if err != nil {
		// Even if no session, clear cookie and return success
		auth.ClearSessionCookie(w)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Logged out successfully"))
		return
	}

	// Invalidate session in database
	if err := auth.InvalidateSession(h.db, sessionID); err != nil {
		log.Printf("Failed to invalidate session: %v", err)
	}

	// Clear cookie
	auth.ClearSessionCookie(w)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}

// Helper functions

func (h *AuthHandler) getUserByID(userID []byte) (*models.User, error) {
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

func (h *AuthHandler) userToResponse(user *models.User) models.UserResponse {
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
