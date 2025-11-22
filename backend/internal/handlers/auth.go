package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"social-network/app"
	"social-network/internal/middleware"
	"social-network/internal/models"
	db_users "social-network/pkg/db/queries/users"
	"social-network/pkg/db/sqlite"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration
func Register(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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

		//insert user
		transaction, err := sqlite.NewQuery(app.DB).Users.CreateUser(r.Context(), db_users.CreateUserParams{
			UserID:       userID,
			Email:        req.Email,
			PasswordHash: hashedPassword,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Dob:          req.DOB,
			Avatar:       sql.NullString{String: req.Avatar, Valid: true},
			Nickname:     sql.NullString{String: req.Nickname, Valid: true},
			AboutMe:      sql.NullString{String: req.AboutMe, Valid: true},
		})

		if err != nil {
			app.Logger.Error("failed to create user")
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		// Create session for auto-login
		sessionID, err := middleware.CreateSession(app.DB, transaction.UserID)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			http.Error(w, "User created but failed to create session", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		middleware.SetSessionCookie(w, sessionID)

		// return user response
		response := models.UserResponse{
			UserID:    hex.EncodeToString(transaction.UserID),
			Email:     transaction.Email,
			FirstName: transaction.FirstName,
			LastName:  transaction.LastName,
			DOB:       transaction.Dob,
			Avatar:    transaction.Avatar.String,
			Nickname:  transaction.Nickname.String,
			AboutMe:   transaction.AboutMe.String,
			CreatedAt: transaction.CreatedAt.Time.String(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}

}

// Login handles user login
func Login(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
		transaction, err := sqlite.NewQuery(app.DB).Users.GetUserByEmail(r.Context(), req.Email)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		// Compare password with bcrypt
		if err := bcrypt.CompareHashAndPassword(transaction.PasswordHash, []byte(req.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Create session
		sessionID, err := middleware.CreateSession(app.DB, transaction.UserID)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		middleware.SetSessionCookie(w, sessionID)

		// Return user response
		response := models.UserResponse{
			UserID:    hex.EncodeToString(transaction.UserID),
			Email:     transaction.Email,
			FirstName: transaction.FirstName,
			LastName:  transaction.LastName,
			DOB:       transaction.Dob,
			Avatar:    transaction.Avatar.String,
			Nickname:  transaction.Nickname.String,
			AboutMe:   transaction.AboutMe.String,
			CreatedAt: transaction.CreatedAt.Time.String(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// GetSession returns the current user from session cookie (protected endpoint)
func GetSession(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get user_id from context (set by auth middleware)
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch user
		user, err := sqlite.NewQuery(app.DB).Users.GerUserById(r.Context(), userID)
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
		response := models.UserResponse{
			UserID:    hex.EncodeToString(user.UserID),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			DOB:       user.Dob,
			Avatar:    user.Avatar.String,
			Nickname:  user.Nickname.String,
			AboutMe:   user.AboutMe.String,
			CreatedAt: user.CreatedAt.Time.String(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

}

// Logout handles user logout
func Logout(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get session cookie
		sessionID, err := middleware.GetSessionCookie(r)
		if err != nil {
			// Even if no session, clear cookie and return success
			middleware.ClearSessionCookie(w)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Logged out successfully"))
			return
		}

		// Invalidate session in database
		if err := middleware.InvalidateSession(app.DB, sessionID); err != nil {
			log.Printf("Failed to invalidate session: %v", err)
		}

		// Clear cookie
		middleware.ClearSessionCookie(w)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Logged out successfully"))
	}

}
