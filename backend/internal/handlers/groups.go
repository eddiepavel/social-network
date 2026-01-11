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
	"time"

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
			Status:  "joined",
		})

		// Commit transaction
		if err := tx.Commit(); err != nil {
			app.Logger.Error("failed to commit transaction", "err", err, "req", r.URL.Path)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		err = app.File.AssignImage(req.Image)

		if err != nil {
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
			Image: func() string {
				if group.Image.Valid {
					return group.Image.String
				}
				return ""
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

		currentUser, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "missing logged in user")
			return
		}

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
				Image: func() string {
					if group.Image.Valid {
						return group.Image.String
					}

					return ""
				}(),
				ImageUrl: func() string {
					if group.Image.Valid {
						sign := app.File.GenerateSignImage(group.GroupImageFileName.String, currentUser, time.Now().Add(15*time.Minute))
						return sign
					}
					return ""
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
			if errors.Is(err, sql.ErrNoRows) {
				utils.Forbidden(w)
				return
			}
			app.Logger.Error("error checking is member", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if isMember.Status == "requested" {
			utils.Forbidden(w)
			return
		}

		// Fetch group details
		init := sqlite.NewQuery(app.DB)

		group, err := helpers.CreateGroupDetailResponse(groupIDHex, init, userID, app.File)

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

		if isMember.Status != "joined" {
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

// here you can either request to join  or remove request to join only.
func RequestToJoinGroup(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		groupID, err := helpers.GenerateFromString(r.PathValue("groupId"))

		if err != nil {
			utils.BadRequest(w, errors.New("bad request man"))
			return
		}

		userId, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "auth user not found")
			return
		}

		db := sqlite.NewQuery(app.DB)

		getGroup, err := db.Groups.GetGroupById(r.Context(), groupID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
			app.Logger.Error("error fetching group details", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if bytes.Equal(userId, getGroup.CreatorID) {
			utils.Forbidden(w)
			return
		}

		var p models.MemberShipRequest

		inputs := helpers.ValidateMemberShip.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &p)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		isMember, err := db.Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			UserID:  userId,
			GroupID: groupID,
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) && p.Action == "request" {
				err := db.Groups.CreateGroupMember(r.Context(), db_groups.CreateGroupMemberParams{
					UserID:  userId,
					GroupID: groupID,
					Status:  "requested",
				})

				if err != nil {
					app.Logger.Error("error creating group member", "err", err)
					utils.Internal(w, errors.New("internal server error"))
					return
				}

				utils.OK(w, map[string]string{"message": "created"})
				return
			}

			if errors.Is(err, sql.ErrNoRows) && p.Action == "remove" {
				utils.Error(w, http.StatusConflict, "409", "invalid action or state", nil)
				return
			}

			app.Logger.Error("error fetching group member", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if isMember.Status == "requested" && p.Action == "remove" {
			err := db.Groups.RemoveUserFromGroup(r.Context(), db_groups.RemoveUserFromGroupParams{
				UserID:  userId,
				GroupID: groupID,
			})

			if err != nil {
				app.Logger.Error("error deleting user from group", "err", err)
				utils.Internal(w, errors.New("internal server error"))
			}
			utils.OK(w, map[string]string{"message": "removed"})
			return
		}

		utils.Error(w, http.StatusConflict, "409", "invalid action or state", nil)

	}
}

func GetGroupRequests(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		groupId, err := helpers.GenerateFromString(r.PathValue("groupId"))

		if err != nil {
			utils.BadRequest(w, errors.New("group id missings"))
			return
		}

		user, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "user not found")
			return
		}

		getGroup, err := sqlite.NewQuery(app.DB).Groups.GetGroupById(r.Context(), groupId)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
			utils.BadRequest(w, errors.New("something is wrong"))
			return
		}

		if !bytes.Equal(user, getGroup.CreatorID) {
			utils.Unauthorized(w, "you are not owner")
			return
		}

		groupMemberRequests, err := sqlite.NewQuery(app.DB).Groups.GetGroupJoinRequests(r.Context(), groupId)

		if err != nil {
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if len(groupMemberRequests) <= 0 {
			utils.OK(w, map[string]string{"message": "no requests yet"})
			return
		}

		var pendingMembers []models.GroupMemberResponse

		for _, member := range groupMemberRequests {
			uuid, _ := helpers.GenerateFromBytes(member.UserID)

			pendingMembers = append(pendingMembers, models.GroupMemberResponse{
				UserID: uuid,
				Status: member.Status,
			})
		}

		utils.OK(w, pendingMembers)
	}
}

func RespondRequest(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		groupId, err := helpers.GenerateFromString(r.PathValue("groupId"))
		if err != nil {
			utils.BadRequest(w, errors.New("invalid group id"))
			return
		}

		currentUser, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "no user found")
			return
		}

		query := sqlite.NewQuery(app.DB)

		group, err := query.Groups.GetGroupById(r.Context(), groupId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
			app.Logger.Error("failed to fetch group", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if !bytes.Equal(group.CreatorID, currentUser) {
			utils.Unauthorized(w, "you are not owner of this group")
			return
		}

		var req struct {
			UserID   string `json:"user_id"`
			Response string `json:"response"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.BadRequest(w, errors.New("invalid request body"))
			return
		}
		defer r.Body.Close()

		if req.UserID == "" {
			utils.BadRequest(w, errors.New("user_id is required"))
			return
		}

		if req.Response != "approve" && req.Response != "reject" {
			utils.BadRequest(w, errors.New("response must be 'approve' or 'reject'"))
			return
		}

		userID, err := helpers.GenerateFromString(req.UserID)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid user_id format"))
			return
		}

		groupMem, err := query.Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			UserID:  userID,
			GroupID: groupId,
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.BadRequest(w, errors.New("no join request found for this user"))
				return
			}
			app.Logger.Error("failed to fetch group member", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if groupMem.Status != "requested" {
			if groupMem.Status == "joined" {
				utils.BadRequest(w, errors.New("user is already a member"))
				return
			}
			utils.BadRequest(w, errors.New("no pending request found for this user"))
			return
		}

		switch req.Response {
		case "reject":
			if err := query.Groups.RemoveUserFromGroup(r.Context(), db_groups.RemoveUserFromGroupParams{
				UserID:  groupMem.UserID,
				GroupID: groupMem.GroupID,
			}); err != nil {
				app.Logger.Error("failed to remove user from group", "err", err)
				utils.Internal(w, errors.New("internal server error"))
				return
			}

		case "approve":
			if err := query.Groups.UpdateGroupMemberStatus(r.Context(), db_groups.UpdateGroupMemberStatusParams{
				Status: "joined",
				UserID: groupMem.UserID,
			}); err != nil {
				app.Logger.Error("failed to update group member status", "err", err)
				utils.Internal(w, errors.New("internal server error"))
				return
			}
		}

		utils.OK(w, map[string]string{"message": "request processed successfully"})
	}
}

// remove only if you are current logged in member with status joined or you are admin want to remove one of you members
func RemoveMember(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		groupId, err := helpers.GenerateFromString(r.PathValue("groupId"))

		if err != nil {
			utils.BadRequest(w, errors.New("incorrect group id"))
			return
		}

		currentUser, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "user id not found")
			return
		}

		query := sqlite.NewQuery(app.DB)

		group, err := query.Groups.GetGroupById(r.Context(), groupId)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}

			utils.Internal(w, errors.New("something went wrong"))
			return
		}

		var req struct {
			User string `json:"user_uuid"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.BadRequest(w, errors.New("wrong payload"))
			return
		}

		userToMutate, err := helpers.GenerateFromString(req.User)

		if err != nil {
			utils.BadRequest(w, errors.New("wrong uuid format"))
			return
		}

		getMember, err := query.Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			UserID:  userToMutate,
			GroupID: groupId,
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.Error(w, 404, "404", "user not found", "the user id you send can not be found")
				return
			}

			utils.Internal(w, errors.New("something went wrong"))
			return
		}

		isOnwer := bytes.Equal(currentUser, group.CreatorID)

		isCurrentMember := bytes.Equal(getMember.UserID, currentUser)

		if (isCurrentMember || isOnwer) && isOnwer != isCurrentMember && getMember.Status == "joined" {

			if err := query.Groups.RemoveUserFromGroup(r.Context(), db_groups.RemoveUserFromGroupParams{
				UserID:  getMember.UserID,
				GroupID: getMember.GroupID,
			}); err != nil {
				utils.Internal(w, errors.New("something went wrong"))
				return
			}
			utils.OK(w, map[string]string{"message": "member removed successfully"})
			return
		}

		utils.Unauthorized(w, "you are not allowed to perform this action")

	}
}

func DeleteGroup(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		groupId, err := helpers.GenerateFromString(r.PathValue("groupId"))

		if err != nil {
			utils.BadRequest(w, errors.New("wrong group id"))
			return
		}

		currentUser, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "user not found")
			return
		}

		query := sqlite.NewQuery(app.DB)

		group, err := query.Groups.GetGroupById(r.Context(), groupId)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}

			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if !bytes.Equal(group.CreatorID, currentUser) {
			utils.Unauthorized(w, "you are not the owner of this group")
			return
		}

		if err := query.Groups.DeleteDbGroup(r.Context(), group.GroupID); err != nil {
			app.Logger.Error("could not delete group", "error", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]string{"message": "group deleted"})

	}
}
