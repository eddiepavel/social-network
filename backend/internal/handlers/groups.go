package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_groups "social-network/pkg/db/queries/groups"
	"social-network/pkg/db/sqlite"

	"github.com/google/uuid"
)

// CreateGroup handles POST /api/groups
// Creates a new group and automatically adds creator as member with status 'joined'
func CreateGroup(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var req models.CreateGroupRequest

		inputs := helpers.ValidateCreateGroup.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		// Generate UUID for group
		groupUUID, err := uuid.New().MarshalBinary()
		if err != nil {
			app.Logger.Error("failed to create marshal binary", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Begin transaction
		tx, err := app.DB.Begin()
		if err != nil {
			app.Logger.Error("failed to begin transaction", "err", err, "req", r.URL.Path)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		defer func() {
			if err != nil {
				tx.Rollback()
			}
		}()

		group, err := sqlite.NewQuery(app.DB).Groups.WithTx(tx).CreateGroup(r.Context(), db_groups.CreateGroupParams{
			GroupID:     groupUUID,
			GroupName:   req.GroupName,
			Description: req.Description,
			Image: func() sql.NullString {
				if req.Image != "" {
					return sql.NullString{Valid: true, String: req.Image}
				}
				return sql.NullString{Valid: false, String: ""}
			}(),
			CreatorID: userID,
		})

		if err != nil {
			app.Logger.Error("failed to begin transaction", "err", err, "req", r.URL.Path)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		err = sqlite.NewQuery(app.DB).Groups.WithTx(tx).CreateGroupMember(r.Context(), db_groups.CreateGroupMemberParams{
			UserID:  userID,
			GroupID: group.GroupID,
		})

		// Commit transaction
		if err := tx.Commit(); err != nil {
			app.Logger.Error("failed to commit transaction", "err", err, "req", r.URL.Path)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		setUUid, _ := helpers.GenerateFromBytes(groupUUID)
		setUserUuid, _ := helpers.GenerateFromBytes(group.CreatorID)
		// Return created group
		response := models.GroupResponse{
			GroupID:     setUUid,
			GroupName:   req.GroupName,
			Description: req.Description,
			Image: func() *string {
				if group.Image.Valid {
					return &group.Image.String
				}

				return nil
			}(),
			CreatorID: setUserUuid,
			CreatedAt: group.CreatedAt.String(),
		}

		utils.OK(w, response)
	}
}

// GetGroups handles GET /api/groups
// Returns all groups with member count
func GetGroups(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		groups, err := sqlite.NewQuery(app.DB).Groups.GetGroupsWithMemberCount(r.Context())

		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				utils.OK(w, []models.GroupResponse{})
				return
			}

			app.Logger.Error("failed to commit transaction", "err", err, "req", r.URL.Path)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		var groupCollection []models.GroupResponse
		for _, group := range groups {

			setGroupUUid, _ := helpers.GenerateFromBytes(group.GroupID)
			setCreatorId, _ := helpers.GenerateFromBytes(group.CreatorID)

			groupCollection = append(groupCollection, models.GroupResponse{
				GroupID:     setGroupUUid,
				GroupName:   group.GroupName,
				Description: group.GroupName,
				Image: func() *string {
					if group.Image.Valid {
						return &group.Image.String
					}

					return nil
				}(),
				CreatorID:   setCreatorId,
				CreatedAt:   group.CreatedAt.String(),
				MemberCount: group.MemberCount,
			})
		}

		utils.OK(w, groupCollection)
	}
}

// GetGroupDetails handles GET /api/groups/:id
// Returns group details with members list (only accessible to members with status 'joined')
func GetGroup(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Extract group ID from URL
		groupIDHex, err := helpers.GenerateFromString(r.PathValue("groupId"))
		if err != nil {
			utils.BadRequest(w, errors.New("group id missing from path"))
			return
		}

		// Check if user is a member with status 'joined'
		isMember, err := sqlite.NewQuery(app.DB).Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			GroupID: groupIDHex,
			UserID:  userID,
		})

		if err != nil {
			app.Logger.Error("error checking is member", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if isMember == 0 {
			utils.Forbidden(w)
			return
		}

		// Fetch group details
		init := sqlite.NewQuery(app.DB)

		group, err := helpers.CreateGroupDetailResponse(groupIDHex, init)

		if err != nil {
			app.Logger.Error("error fetching group details", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, group)
	}
}

// invite users to a group
func InviteToGroup(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userId, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "unauthorized")
		}

		groupID, err := helpers.GenerateFromString(r.PathValue("groupId"))

		if err != nil {
			utils.BadRequest(w, errors.New("bad request man"))
			return
		}

		db := sqlite.NewQuery(app.DB)

		isMember, _ := db.Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			UserID:  userId,
			GroupID: groupID,
		})

		if isMember == 0 {
			utils.Forbidden(w)
			return
		}

		var p models.InviteGroupRequest

		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			utils.BadRequest(w, err)
			return
		}

		defer r.Body.Close()

		var users [][]byte

		members := make(map[string]db_groups.GroupMember)

		groupMember, _ := db.Groups.GetGroupMembersWithRequests(r.Context(), groupID)

		for _, member := range groupMember {
			id, _ := helpers.GenerateFromBytes(member.UserID)
			members[id] = db_groups.GroupMember{
				Status:    member.Status,
				InvitedBy: member.InvitedBy,
			}
		}

		//users that have already request to join we instantly accept them
		//this is wrong in so many levels but im bored to create new solution for already requested users. Please dont judge
		//better approach is to collect users that are status requested and pass them in UPDATE WHERE IN
		for _, id := range p.Users {
			gen, _ := helpers.GenerateFromString(id)
			member, exists := members[id]

			if !exists && !bytes.Equal(userId, gen) {
				users = append(users, gen)
				continue
			}

			if exists && member.Status == "requested" && len(member.InvitedBy) == 0 {
				err := db.Groups.UpdateGroupMemberStatus(r.Context(), db_groups.UpdateGroupMemberStatusParams{
					Status: "joined",
					UserID: gen,
				})

				if err != nil {
					app.Logger.Error("failed to update status", "user", id)
				}
			}

		}
		//validate the users we have collected from the payload in order to create
		check, err := db.Users.ValidateUserIds(r.Context(), users)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
			utils.Internal(w, errors.New(err.Error()))
			return
		}
		//validate if everyid is actually legit
		if int(check) != len(users) {
			utils.NotFound(w)
			return
		}

		//handle if all users are already invited
		if len(users) == 0 {
			utils.OK(w, map[string]string{"message": "users already invited"})
			return
		}

		tx, _ := app.DB.Begin()

		defer tx.Rollback()

		for i := range users {
			err := db.Groups.WithTx(tx).InviteGroupMembers(r.Context(), db_groups.InviteGroupMembersParams{
				UserID:    users[i],
				GroupID:   groupID,
				Status:    "requested",
				InvitedBy: userId,
			})

			if err != nil {
				app.Logger.Error("failed to invite user", "err", err)
				utils.Internal(w, errors.New("failed to send invitations"))
				return
			}
		}

		if err := tx.Commit(); err != nil {
			app.Logger.Error("failed to commit transaction", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]interface{}{"message": "Invitations sent successfully", "invited": len(users)})

	}
}

// // InviteUser handles POST /api/groups/:id/invite
// // Invites a user to join the group (requester must be a member)
// func (h *GroupsHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Get current user from context
// 	inviterID, ok := middleware.GetUserIDFromContext(r.Context())
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Extract group ID from URL
// 	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/groups/"), "/")
// 	if len(pathParts) < 2 || pathParts[0] == "" {
// 		http.Error(w, "Group ID required", http.StatusBadRequest)
// 		return
// 	}

// 	groupIDHex := pathParts[0]
// 	groupID, err := hex.DecodeString(groupIDHex)
// 	if err != nil {
// 		http.Error(w, "Invalid group ID format", http.StatusBadRequest)
// 		return
// 	}

// 	// Parse request body
// 	var req models.InviteUserRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Validate user_id
// 	if req.UserID == "" {
// 		http.Error(w, "user_id is required", http.StatusBadRequest)
// 		return
// 	}

// 	inviteeID, err := hex.DecodeString(req.UserID)
// 	if err != nil {
// 		http.Error(w, "Invalid user_id format", http.StatusBadRequest)
// 		return
// 	}

// 	// Check if inviter is a member with status 'joined'
// 	isMember, err := h.isGroupMember(inviterID, groupID)
// 	if err != nil {
// 		log.Printf("Error checking membership: %v", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	if !isMember {
// 		http.Error(w, "Only members can invite users", http.StatusForbidden)
// 		return
// 	}

// 	// Check if invitee exists
// 	var exists bool
// 	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = ?)", inviteeID).Scan(&exists)
// 	if err != nil {
// 		log.Printf("Error checking user existence: %v", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
// 	if !exists {
// 		http.Error(w, "User not found", http.StatusNotFound)
// 		return
// 	}

// 	// Check if invitee is already a member or has pending request
// 	var existingStatus *string
// 	err = h.db.QueryRow(
// 		"SELECT status FROM group_members WHERE user_id = ? AND group_id = ?",
// 		inviteeID, groupID,
// 	).Scan(&existingStatus)

// 	if err != nil && err != sql.ErrNoRows {
// 		log.Printf("Error checking existing membership: %v", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	if existingStatus != nil {
// 		if *existingStatus == "joined" {
// 			http.Error(w, "User is already a member", http.StatusConflict)
// 			return
// 		}
// 		http.Error(w, "User already has a pending invitation or request", http.StatusConflict)
// 		return
// 	}

// 	// Insert invitation with status 'requested'
// 	insertQuery := `
// 		INSERT INTO group_members (user_id, group_id, status, invited_by, created_at)
// 		VALUES (?, ?, 'requested', ?, ?)
// 	`
// 	now := time.Now()
// 	_, err = h.db.Exec(insertQuery, inviteeID, groupID, inviterID, now)
// 	if err != nil {
// 		log.Printf("Failed to create invitation: %v", err)
// 		http.Error(w, "Failed to send invitation", http.StatusInternalServerError)
// 		return
// 	}

// 	// TODO: Create notification for invitee

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{
// 		"message": "Invitation sent successfully",
// 	})
// }

// // RequestToJoin handles POST /api/groups/:id/request
// // User requests to join a group
// func (h *GroupsHandler) RequestToJoin(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Get current user from context
// 	userID, ok := middleware.GetUserIDFromContext(r.Context())
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Extract group ID from URL
// 	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/groups/"), "/")
// 	if len(pathParts) < 2 || pathParts[0] == "" {
// 		http.Error(w, "Group ID required", http.StatusBadRequest)
// 		return
// 	}

// 	groupIDHex := pathParts[0]
// 	groupID, err := hex.DecodeString(groupIDHex)
// 	if err != nil {
// 		http.Error(w, "Invalid group ID format", http.StatusBadRequest)
// 		return
// 	}

// 	// Check if group exists
// 	var exists bool
// 	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM groups WHERE group_id = ?)", groupID).Scan(&exists)
// 	if err != nil {
// 		log.Printf("Error checking group existence: %v", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
// 	if !exists {
// 		http.Error(w, "Group not found", http.StatusNotFound)
// 		return
// 	}

// 	// Check if user already has a membership or pending request
// 	var existingStatus *string
// 	err = h.db.QueryRow(
// 		"SELECT status FROM group_members WHERE user_id = ? AND group_id = ?",
// 		userID, groupID,
// 	).Scan(&existingStatus)

// 	if err != nil && err != sql.ErrNoRows {
// 		log.Printf("Error checking existing membership: %v", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	if existingStatus != nil {
// 		if *existingStatus == "joined" {
// 			http.Error(w, "You are already a member", http.StatusConflict)
// 			return
// 		}
// 		http.Error(w, "You already have a pending request or invitation", http.StatusConflict)
// 		return
// 	}

// 	// Insert join request with status 'requested'
// 	insertQuery := `
// 		INSERT INTO group_members (user_id, group_id, status, invited_by, created_at)
// 		VALUES (?, ?, 'requested', NULL, ?)
// 	`
// 	now := time.Now()
// 	_, err = h.db.Exec(insertQuery, userID, groupID, now)
// 	if err != nil {
// 		log.Printf("Failed to create join request: %v", err)
// 		http.Error(w, "Failed to send join request", http.StatusInternalServerError)
// 		return
// 	}

// 	// TODO: Create notification for group creator

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{
// 		"message": "Join request sent successfully",
// 	})
// }

// // HandleJoinRequest handles POST /api/groups/:groupId/accept/:userId
// // Creator accepts or rejects join requests
// func (h *GroupsHandler) HandleJoinRequest(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Get current user from context
// 	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Extract group ID and user ID from URL
// 	// URL format: /api/groups/:groupId/accept/:userId
// 	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/groups/"), "/")
// 	if len(pathParts) < 3 || pathParts[0] == "" || pathParts[2] == "" {
// 		http.Error(w, "Group ID and User ID required", http.StatusBadRequest)
// 		return
// 	}

// 	groupIDHex := pathParts[0]
// 	targetUserIDHex := pathParts[2]

// 	groupID, err := hex.DecodeString(groupIDHex)
// 	if err != nil {
// 		http.Error(w, "Invalid group ID format", http.StatusBadRequest)
// 		return
// 	}

// 	targetUserID, err := hex.DecodeString(targetUserIDHex)
// 	if err != nil {
// 		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
// 		return
// 	}

// 	// Parse request body to determine action (accept or reject)
// 	var req models.HandleJoinRequestRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Validate action
// 	if req.Action != "accept" && req.Action != "reject" {
// 		http.Error(w, "action must be 'accept' or 'reject'", http.StatusBadRequest)
// 		return
// 	}

// 	// Check if current user is the group creator
// 	isCreator, err := h.isGroupCreator(currentUserID, groupID)
// 	if err != nil {
// 		log.Printf("Error checking creator status: %v", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	if !isCreator {
// 		http.Error(w, "Only the group creator can accept or reject join requests", http.StatusForbidden)
// 		return
// 	}

// 	// Check if there's a pending request for this user
// 	var currentStatus string
// 	err = h.db.QueryRow(
// 		"SELECT status FROM group_members WHERE user_id = ? AND group_id = ?",
// 		targetUserID, groupID,
// 	).Scan(&currentStatus)

// 	if err == sql.ErrNoRows {
// 		http.Error(w, "No join request found for this user", http.StatusNotFound)
// 		return
// 	}
// 	if err != nil {
// 		log.Printf("Error checking join request: %v", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	if currentStatus != "requested" {
// 		http.Error(w, "No pending join request for this user", http.StatusBadRequest)
// 		return
// 	}

// 	// Update status based on action
// 	newStatus := "joined"
// 	if req.Action == "reject" {
// 		newStatus = "rejected"
// 	}

// 	updateQuery := `
// 		UPDATE group_members
// 		SET status = ?
// 		WHERE user_id = ? AND group_id = ?
// 	`
// 	_, err = h.db.Exec(updateQuery, newStatus, targetUserID, groupID)
// 	if err != nil {
// 		log.Printf("Failed to update join request: %v", err)
// 		http.Error(w, "Failed to process join request", http.StatusInternalServerError)
// 		return
// 	}

// 	// TODO: Create notification for target user

// 	message := "Join request accepted successfully"
// 	if req.Action == "reject" {
// 		message = "Join request rejected successfully"
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{
// 		"message": message,
// 	})
// }
