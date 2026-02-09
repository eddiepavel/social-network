package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/constants"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_groups "social-network/pkg/db/queries/groups"
	db_posts "social-network/pkg/db/queries/posts"
	"social-network/pkg/db/sqlite"
	"strconv"
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
			CreatedAt: time.Now(),
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
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
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
			utils.Error(w, 409, "409", "already processed", "selected users already invited")
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
			// Create notification for group invitation
			err = helpers.CreateNotification(app, users[i], "group_invitation", userId, groupID, nil, tx)

			if err != nil {
				app.Logger.Warn("failed to send notification", "reason", err)
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

				// Create notification for group join request
				groupCreator := getGroup.CreatorID
				err = helpers.CreateNotification(app, groupCreator, constants.NotificationGroupRequest, userId, groupID, nil, nil)
				if err != nil {
					app.Logger.Error("failed to create group request notification", "err", err)
					// Don't fail the request if notification fails
				}

				utils.OK(w, map[string]string{"message": "created"})
				return
			}

			if errors.Is(err, sql.ErrNoRows) && p.Action == "remove" {
				utils.Error(w, http.StatusConflict, "409", "invalid action or state", "invalid action or state")
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

		utils.Error(w, http.StatusConflict, "409", "invalid action or state", "invalid action or state")

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
			utils.OK(w, []models.GroupMemberResponse{})
			return
		}

		var pendingMembers []models.GroupMemberResponse

		for _, member := range groupMemberRequests {
			uuid, _ := helpers.GenerateFromBytes(member.UserID)

			pendingMembers = append(pendingMembers, models.GroupMemberResponse{
				UserID:    uuid,
				Status:    member.Status,
				FirstName: &member.MFirstName,
				LastName:  &member.MLastName,
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

		if req.UserID == "" || req.Response == "" {
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

			// Create notification for group join rejection
			err = helpers.CreateNotification(app, groupMem.UserID, constants.NotificationGroupJoinRejected, currentUser, groupId, nil, nil)
			if err != nil {
				app.Logger.Error("failed to create group join rejected notification", "err", err)
				// Don't fail the request if notification fails
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

			// Create notification for group join approval
			err = helpers.CreateNotification(app, groupMem.UserID, constants.NotificationGroupJoinApproved, currentUser, groupId, nil, nil)
			if err != nil {
				app.Logger.Error("failed to create group join approved notification", "err", err)
				// Don't fail the request if notification fails
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
			User string `json:"user_id"`
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

		if group.Image.Valid {
			err := app.File.RemoveImage(group.GroupImageID.String)

			if err != nil {
				utils.Internal(w, err)
				return
			}
		}

		utils.OK(w, map[string]string{"message": "group deleted"})

	}
}

// GetGroupEvents handles GET /api/groups/{groupId}/events
// Returns all events for a group (members only)
func GetGroupEvents(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		groupID, err := helpers.GenerateFromString(r.PathValue("groupId"))
		if err != nil {
			utils.BadRequest(w, errors.New("invalid group ID"))
			return
		}

		// Check if user is a member
		isMember, err := sqlite.NewQuery(app.DB).Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			GroupID: groupID,
			UserID:  userID,
		})
		if err != nil || isMember.Status != "joined" {
			utils.Forbidden(w)
			return
		}

		events, err := sqlite.NewQuery(app.DB).Groups.GetGroupEvents(r.Context(), db_groups.GetGroupEventsParams{
			GroupID: groupID,
			UserID:  userID,
		})
		if err != nil {
			app.Logger.Error("failed to get group events", "error", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		var eventsList []models.EventResponse
		for _, e := range events {
			eventID, _ := helpers.GenerateFromBytes(e.EventID)
			eventsList = append(eventsList, models.EventResponse{
				EventID:     eventID,
				EventName:   e.Title,
				Description: e.Description,
				Timestamp:   e.EventTimestamp.Format(time.RFC3339),
				CreatedAt: func() time.Time {
					if e.CreatedAt.Valid {
						return e.CreatedAt.Time
					}
					return time.Time{}
				}(),
				GoingCount:    e.GoingCount,
				NotGoingCount: e.NotGoingCount,
				UserRsvp: func() *string {
					if e.UserRsvp.Valid {
						return &e.UserRsvp.String
					}
					return nil
				}(),
			})
		}

		utils.OK(w, eventsList)
	}
}

// CreateGroupEvent handles POST /api/groups/{groupId}/events
// Creates a new event in a group (members only)
func CreateGroupEvent(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		groupID, err := helpers.GenerateFromString(r.PathValue("groupId"))
		if err != nil {
			utils.BadRequest(w, errors.New("invalid group ID"))
			return
		}

		// Check if user is a member
		isMember, err := sqlite.NewQuery(app.DB).Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			GroupID: groupID,
			UserID:  userID,
		})

		if err != nil || isMember.Status != "joined" {
			utils.Forbidden(w)
			return
		}

		inputs := helpers.ValidateEventCreate.Build(r, app)

		var req models.CreateEventRequest

		ok, validationErrors := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, 422, "422", "validation error", validationErrors)
			return
		}

		eventTimestamp, err := time.Parse(time.RFC3339, req.Timestamp)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid timestamp format, use RFC3339"))
			return
		}

		eventID, err := uuid.New().MarshalBinary()
		if err != nil {
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		event, err := sqlite.NewQuery(app.DB).Groups.CreateGroupEvent(r.Context(), db_groups.CreateGroupEventParams{
			EventID:        eventID,
			GroupID:        groupID,
			CreatorID:      userID,
			Title:          req.Title,
			Description:    req.Description,
			EventTimestamp: eventTimestamp,
		})
		if err != nil {
			app.Logger.Error("failed to create event", "error", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Notify all group members about the new event
		members, err := sqlite.NewQuery(app.DB).Groups.GetGroupMemberIDs(r.Context(), groupID)
		if err == nil {
			for _, memberID := range members {
				if !bytes.Equal(memberID, userID) {
					_ = helpers.CreateNotification(app, memberID, "group_event", userID, groupID, event.EventID, nil)
				}
			}
		}

		eventIDStr, _ := helpers.GenerateFromBytes(event.EventID)
		response := models.EventResponse{
			EventID:     eventIDStr,
			EventName:   event.Title,
			Description: event.Description,
			Timestamp:   event.EventTimestamp.Format(time.RFC3339),
			CreatedAt: func() time.Time {
				if event.CreatedAt.Valid {
					return event.CreatedAt.Time
				}
				return time.Now()
			}(),
			GoingCount:    0,
			NotGoingCount: 0,
		}

		utils.OK(w, response)
	}
}

// RSVPToEvent handles POST /api/events/{eventId}/rsvp
// Updates user's RSVP status for an event
func RSVPToEvent(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		eventID, err := helpers.GenerateFromString(r.PathValue("eventId"))
		if err != nil {
			utils.BadRequest(w, errors.New("invalid event ID"))
			return
		}

		// Get event's group ID to verify membership
		groupID, err := sqlite.NewQuery(app.DB).Groups.GetEventGroupID(r.Context(), eventID)
		if err != nil {
			utils.NotFound(w)
			return
		}

		// Check if user is a group member
		isMember, err := sqlite.NewQuery(app.DB).Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			GroupID: groupID,
			UserID:  userID,
		})
		if err != nil || isMember.Status != "joined" {
			utils.Forbidden(w)
			return
		}

		var req models.RSVPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.BadRequest(w, errors.New("invalid request body"))
			return
		}
		defer r.Body.Close()

		// Validate status
		validStatuses := map[string]bool{"going": true, "not going": true, "maybe": true}
		if !validStatuses[req.Status] {
			utils.BadRequest(w, errors.New("status must be 'going', 'not going', or 'maybe'"))
			return
		}

		eventIDStr, _ := helpers.GenerateFromBytes(eventID)
		err = sqlite.NewQuery(app.DB).Groups.UpsertRSVP(r.Context(), db_groups.UpsertRSVPParams{
			EventID: eventIDStr,
			UserID:  userID,
			Status:  sql.NullString{String: req.Status, Valid: true},
		})
		if err != nil {
			app.Logger.Error("failed to upsert RSVP", "error", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]string{"message": "RSVP updated", "status": req.Status})
	}
}

// GetGroupPosts handles GET /api/groups/{groupId}/posts
// Returns all posts in a group (members only)
func GetGroupPosts(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		groupID, err := helpers.GenerateFromString(r.PathValue("groupId"))
		if err != nil {
			utils.BadRequest(w, errors.New("invalid group ID"))
			return
		}

		// Check if user is a member
		isMember, err := sqlite.NewQuery(app.DB).Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			GroupID: groupID,
			UserID:  userID,
		})
		if err != nil || isMember.Status != "joined" {
			utils.Forbidden(w)
			return
		}

		// Parse pagination parameters
		page := 1
		size := 10

		if pageParam := r.URL.Query().Get("page"); pageParam != "" {
			if parsedPage, err := strconv.Atoi(pageParam); err == nil && parsedPage > 0 {
				page = parsedPage
			}
		}

		if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
			if parsedSize, err := strconv.Atoi(sizeParam); err == nil && parsedSize > 0 {
				size = parsedSize
			}
		}

		offset := int64((page - 1) * size)
		limit := int64(size)

		posts, err := sqlite.NewQuery(app.DB).Posts.GetGroupPosts(r.Context(), db_posts.GetGroupPostsParams{
			AuthorID: userID,
			GroupID:  groupID,
			Limit:    limit,
			Offset:   offset,
		})
		if err != nil {
			app.Logger.Error("failed to get group posts", "error", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		var postsList []models.FeedPostResponse
		for _, p := range posts {
			postID, _ := helpers.GenerateFromBytes(p.PostID)
			authorID, _ := helpers.GenerateFromBytes(p.AuthorID)

			postsList = append(postsList, models.FeedPostResponse{
				PostID:   postID,
				AuthorID: authorID,
				Content:  p.Content,
				ImageID: func() *string {
					if p.ImageID.Valid {
						return &p.ImageID.String
					}
					return nil
				}(),
				ImageUrl: func() string {
					if p.ImageID.Valid && p.FileName.Valid {
						return app.File.GenerateSignImage(p.FileName.String, userID, time.Now().Add(15*time.Minute))
					}
					return ""
				}(),
				Visibility: p.Visibility,
				CreatedAt: func() time.Time {
					if p.CreatedAt.Valid {
						return p.CreatedAt.Time
					}
					return time.Time{}
				}(),
				UserReacted:   p.UserReacted != 0,
				ReactionCount: p.ReactionCount,
				CommentCount:  p.CommentCount,
				AuthorName:    p.FirstName + " " + p.LastName,
				AuthorAvatar: func() *string {
					if p.Avatar.Valid {
						return &p.Avatar.String
					}
					return nil
				}(),
			})
		}

		utils.OK(w, postsList)
	}
}

// CreateGroupPost handles POST /api/groups/{groupId}/posts
// Creates a new post in a group (members only)
func CreateGroupPost(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		groupID, err := helpers.GenerateFromString(r.PathValue("groupId"))
		if err != nil {
			utils.BadRequest(w, errors.New("invalid group ID"))
			return
		}

		// Check if user is a member
		isMember, err := sqlite.NewQuery(app.DB).Groups.IsGroupMember(r.Context(), db_groups.IsGroupMemberParams{
			GroupID: groupID,
			UserID:  userID,
		})
		if err != nil || isMember.Status != "joined" {
			utils.Forbidden(w)
			return
		}

		var req models.CreatePostRequest
		inputs := helpers.ValidatePost.Build(r, app)
		ok, errValidation := utils.Validate(r, inputs, &req)
		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		postID, err := uuid.New().MarshalBinary()
		if err != nil {
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		var image sql.NullString
		if req.ImageID != "" {
			image = sql.NullString{String: req.ImageID, Valid: true}
		}

		post, err := sqlite.NewQuery(app.DB).Posts.CreateGroupPost(r.Context(), db_posts.CreateGroupPostParams{
			PostID:   postID,
			AuthorID: userID,
			Content:  req.Content,
			ImageID:  image,
			GroupID:  groupID,
		})
		if err != nil {
			app.Logger.Error("failed to create group post", "error", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		if image.Valid {
			_ = app.File.AssignImage(image.String)
		}

		postIDStr, _ := helpers.GenerateFromBytes(post.PostID)
		authorIDStr, _ := helpers.GenerateFromBytes(post.AuthorID)

		response := models.PostResponse{
			PostID:     postIDStr,
			AuthorID:   post.AuthorID,
			Content:    post.Content,
			Visibility: post.Visibility,
			CreatedAt: func() time.Time {
				if post.CreatedAt.Valid {
					return post.CreatedAt.Time
				}
				return time.Now()
			}(),
			ImageID: image.String,
		}
		_ = authorIDStr

		utils.OK(w, response)
	}
}

func UpdateGroup(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		groupId, err := helpers.GenerateFromString(r.PathValue("groupId"))

		if err != nil {
			utils.BadRequest(w, errors.New("wrong group id"))
			return
		}

		inputs := helpers.ValidateUpdateGroup.Build(r, app)

		req := models.CreateGroupRequest{}

		ok, errValidation := utils.Validate(r, inputs, &req)

		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
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

		if !bytes.Equal(currentUser, group.CreatorID) {
			utils.Unauthorized(w, "you are not allowed to do moditifications to this group")
			return
		}

		var image sql.NullString

		if req.Image != "" {
			if group.Image.Valid && req.Image != group.Image.String {

				err := app.File.RemoveImage(group.GroupImageID.String)

				if err != nil {
					utils.Internal(w, errors.New("failed to update something went wrong"))
					return
				}

				err = app.File.AssignImage(req.Image)

				if err != nil {
					utils.Internal(w, errors.New("failed to update something went wrong"))
					return
				}

			} else if !group.Image.Valid {
				err = app.File.AssignImage(req.Image)

				if err != nil {
					utils.Internal(w, errors.New("failed to update something went wrong"))
					return
				}
			}

			image = sql.NullString{Valid: true, String: req.Image}

		} else if req.Image == "" && group.Image.Valid {

			image = sql.NullString{Valid: true, String: group.Image.String}
		}

		updateGroup, err := query.Groups.UpdateDbGroup(r.Context(), db_groups.UpdateDbGroupParams{
			GroupName:   req.GroupName,
			Description: req.Description,
			Image:       image,
			GroupID:     groupId,
			CreatorID:   currentUser,
		})

		if err != nil {
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		response := models.GroupResponse{
			GroupID:     r.PathValue("groupId"),
			GroupName:   updateGroup.GroupName,
			Description: updateGroup.Description,
			Image: func() string {
				if group.Image.Valid {
					return updateGroup.Image.String
				}
				return ""
			}(),
		}

		utils.OK(w, response)

	}
}
