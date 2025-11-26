package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_users "social-network/pkg/db/queries/users"
	"social-network/pkg/db/sqlite"
)

// GetUserProfile handles GET /api/users/:id
// Returns user profile respecting privacy settings
func GetUserProfile(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Extract user ID from URL path: /api/users/:id
		targetUserIDHex := r.PathValue("id")
		if targetUserIDHex == "" {
			utils.BadRequest(w, errors.New("user ID required"))
			return
		}

		targetUserID, err := helpers.GenerateFromString(targetUserIDHex)
		if err != nil {
			utils.BadRequest(w, err)
			return
		}

		// Fetch user profile
		user := helpers.FetchUser(app, targetUserID, r.Context(), w)
		if user.UserID == nil {
			return
		}

		// Check if user can view this profile
		access, err := helpers.AccessToProfile(currentUserID, user, app, r)

		if !access && errors.Is(err, sql.ErrNoRows) {
			app.Logger.Error("Error checking profile access", "error", err)
			utils.Unauthorized(w, "you have no access to view this user profile")
			return
		}

		if err != nil {
			app.Logger.Error("Error checking profile access", "error", err)
			utils.Internal(w, err)
			return
		}
		// Return user response
		response := helpers.UserToResponse(user)

		utils.Write(w, http.StatusOK, response)
	}

}

// UpdateProfile handles PUT /api/users/profile
// Updates current user's profile (own profile only)
func UpdateProfile(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		user := helpers.FetchUser(app, userID, r.Context(), w)
		if user.UserID == nil {
			return // Error already handled in FetchUser
		}

		// Parse request body
		var req models.UpdateProfileRequest

		inputs := helpers.ValidateUpdateProfile.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusBadRequest, "400", "validation error", errValidation)
			return
		}

		user = helpers.UpdateUser(user, req)
		if user.UserID == nil {
			utils.BadRequest(w, errors.New("no new data provided"))
			return
		}

		updated, err := sqlite.NewQuery(app.DB).Users.UpdateUser(r.Context(), db_users.UpdateUserParams{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Nickname:  user.Nickname,
			AboutMe:   user.AboutMe,
			Avatar:    user.Avatar,
			UserID:    user.UserID,
		})
		if err != nil {
			app.Logger.Error("Failed to update user", "error", err.Error())
			utils.Internal(w, err)
			return
		}

		// Return user response
		response := helpers.UserToResponse(updated)

		utils.Write(w, http.StatusOK, response)
	}
}

// UpdatePrivacy handles PUT /api/users/privacy
// Updates current user's privacy settings (own profile only)
func UpdatePrivacy(app *app.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Parse request body
		var req models.UpdatePrivacyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.BadRequest(w, err)
			return
		}

		// Update privacy setting
		user, err := sqlite.NewQuery(app.DB).Users.UpdateUserPrivacy(r.Context(), db_users.UpdateUserPrivacyParams{IsPublic: req.IsPublic, UserID: userID})
		if err == sql.ErrNoRows {
			utils.NotFound(w)
			return
		}
		if err != nil {
			app.Logger.Error("Failed to fetch user", "error", err.Error())
			utils.Internal(w, err)
			return
		}

		// Return updated user
		response := helpers.UserToResponse(user)

		utils.Write(w, http.StatusOK, response)
	}
}
