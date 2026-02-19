package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_users "social-network/pkg/db/queries/users"
	"social-network/pkg/db/sqlite"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration
func Register(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Validate required fields

		var req models.CreateUserRequest

		inputs := helpers.ValidateRegister.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		// Generate UUID for user
		userID, err := uuid.New().MarshalBinary()
		if err != nil {
			app.Logger.Error("failed uuid", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Hash password with bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			app.Logger.Error("failed to hash password", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
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
			app.Logger.Error("failed to create user", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Create session for auto-login
		session, err := middleware.CreateSession(app.DB, transaction.UserID)
		if err != nil {
			app.Logger.Error("failed to create session", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Set session cookie
		middleware.SetSessionCookie(w, session.SessionID)

		// return user response
		response := helpers.UserToResponse(transaction)

		utils.OK(w, response)
	}

}

// Login handles user login
func Login(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req models.LoginRequest

		inputs := helpers.ValidateLogin.Build(r, app)

		ok, errorValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errorValidation)
			return
		}

		// Fetch user by email
		transaction, err := sqlite.NewQuery(app.DB).Users.GetUserByEmail(r.Context(), req.Email)

		if err != nil {
			if err == sql.ErrNoRows {
				utils.Unauthorized(w, "Invalid credentials")
				return
			}
			utils.Internal(w, err)
		}

		// Compare password with bcrypt
		if err := bcrypt.CompareHashAndPassword(transaction.PasswordHash, []byte(req.Password)); err != nil {
			utils.Unauthorized(w, "Invalid credentials")
			return
		}

		// Check if session exists for user
		session, err := sqlite.NewQuery(app.DB).Sessions.GetSessionByUserID(r.Context(), transaction.UserID)
		// If no session exists OR session is invalid, create a new one
		if errors.Is(err, sql.ErrNoRows) {
			// No session exists, create new one
			session, err = middleware.CreateSession(app.DB, transaction.UserID)
			if err != nil {
				app.Logger.Error("Failed to create session", "error", err)
				utils.Internal(w, err)
				return
			}
		} else if err != nil {
			// Database error
			utils.Internal(w, err)
			return
		} else {
			// Session exists, validate it
			_, err = middleware.ValidateSession(app.DB, session.SessionID)
			if err != nil {
				// Session exists but is invalid/expired, create new one
				session, err = middleware.CreateSession(app.DB, transaction.UserID)
				if err != nil {
					app.Logger.Error("Failed to create session", "error", err)
					utils.Internal(w, err)
					return
				}
			}
		}
		// Set session cookie
		middleware.SetSessionCookie(w, session.SessionID)

		// Return user response
		response := helpers.UserToResponse(transaction)
		utils.OK(w, response)
	}
}

// GetSession returns the current user from session cookie (protected endpoint)
func GetSession(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get user_id from context (set by auth middleware)
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Fetch user
		user := helpers.FetchUser(app, userID, r.Context(), w)
		if user.UserID == nil {
			return
		}

		// Return user response
		response := helpers.UserToResponseImage(user, app)
		utils.OK(w, response)
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
			utils.OK(w, "Logged out successfully")
			return
		}

		// Invalidate session in database
		if err := middleware.InvalidateSession(app.DB, sessionID); err != nil {
			log.Printf("Failed to invalidate session: %v", err)
		}

		// Clear cookie
		middleware.ClearSessionCookie(w)

		utils.OK(w, "Logged out successfully")
	}

}
