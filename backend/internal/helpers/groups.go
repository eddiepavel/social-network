package helpers

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"social-network/internal/models"
	"social-network/internal/services"
	db_chat "social-network/pkg/db/queries/chat"
	db_groups "social-network/pkg/db/queries/groups"
	"social-network/pkg/db/sqlite"
	"time"
)

func CreateGroupDetailResponse(groupId []byte, t *sqlite.Transactions, user []byte, f *services.FileService, isGroupMember *db_groups.GroupMember) (models.GroupDetailsResponse, error) {

	var group models.GroupDetailsResponse
	ctx := context.Background()

	// Get group
	getGroup, err := t.Groups.GetGroupById(ctx, groupId)
	if err != nil {
		return group, err
	}
	groupUUID, _ := GenerateFromBytes(getGroup.GroupID)

	isOwner := bytes.Equal(getGroup.CreatorID, user)

	getCountMembers, _ := t.Groups.CountMembers(ctx, groupId)

	var memberStatus string

	switch {
	case isGroupMember == nil:
		memberStatus = "not_member"
	case isGroupMember.Status == "requested" && isGroupMember.InvitedBy == nil:
		memberStatus = isGroupMember.Status
	case isGroupMember.Status == "joined":
		memberStatus = isGroupMember.Status
	case isGroupMember.Status == "requested" && len(isGroupMember.InvitedBy) != 0:
		memberStatus = "invited"
	}

	group.Group = models.GroupResponse{
		GroupID:     groupUUID,
		GroupName:   getGroup.GroupName,
		Description: getGroup.Description,
		Image: func() string {
			if getGroup.Image.Valid {
				return getGroup.Image.String
			}
			return ""
		}(),
		ImageUrl: func() string {
			if getGroup.Image.Valid {
				sign := f.GenerateSignImage(getGroup.GroupImageFileName.String, user, time.Now().Add(15*time.Minute))
				return sign
			}
			return ""
		}(),
		CreatedAt:    getGroup.CreatedAt.Local().String(),
		IsOwner:      isOwner,
		TotalMembers: getCountMembers,
		UserStatus:   memberStatus,
	}

	if isGroupMember == nil || isGroupMember.Status == "requested" {
		return group, nil
	}
	// Get members
	getMembers, _ := t.Groups.GetGroupMembers(ctx, groupId)
	for _, member := range getMembers {

		canRemove := func() bool {
			if bytes.Equal(member.UserID, user) && isOwner {
				return false
			}

			return isOwner || bytes.Equal(member.UserID, user)
		}()

		memberUUID, _ := GenerateFromBytes(member.UserID)
		memberResp := models.GroupMemberResponse{
			UserID:    memberUUID,
			Status:    member.Status,
			FirstName: &member.FirstName,
			LastName:  &member.LastName,
			CanRemove: canRemove,
		}
		if member.Avatar.Valid {
			memberResp.Avatar = &member.Avatar.String
		}
		group.Members = append(group.Members, memberResp)
	}

	// Get events with RSVPs and deduplicate
	eventsWithRSVPs, _ := t.Groups.GetGroupEventsWithRSVPs(ctx, groupId)
	eventsMap := make(map[string]*models.EventResponse)

	for _, row := range eventsWithRSVPs {
		eventUUID, _ := GenerateFromBytes(row.EventID)

		// Create event if not exists
		if _, exists := eventsMap[eventUUID]; !exists {
			eventsMap[eventUUID] = &models.EventResponse{
				EventID:     eventUUID,
				EventName:   row.Title,
				Description: row.Description,
				Timestamp:   row.EventTimestamp.String(),
				CreatedAt:   row.EventCreatedAt.Time,
				// CreatedAt:   row.EventCreatedAt.String(),
				RSVPs: []models.RSVPResponse{},
			}
		}

		// Add RSVP if exists
		if row.RsvpUserID != nil {
			rsvpUserUUID, _ := GenerateFromBytes(row.RsvpUserID)
			rsvp := models.RSVPResponse{
				UserID:    rsvpUserUUID,
				FirstName: &row.RsvpFirstName.String,
				LastName:  &row.RsvpLastName.String,
				Status:    row.RsvpStatus.String,
				CreatedAt: row.RsvpCreatedAt.Time.String(),
			}
			if row.RsvpFirstName.Valid {
				rsvp.FirstName = &row.RsvpFirstName.String
			}
			if row.RsvpLastName.Valid {
				rsvp.LastName = &row.RsvpLastName.String
			}
			if row.RsvpAvatar.Valid {
				rsvp.Avatar = &row.RsvpAvatar.String
			}
			eventsMap[eventUUID].RSVPs = append(eventsMap[eventUUID].RSVPs, rsvp)
		}
	}

	// Convert map to slice
	for _, event := range eventsMap {
		group.Events = append(group.Events, *event)
	}

	return group, nil
}

func findGroupChatRoomID(ctx context.Context, db *sql.DB, groupID []byte) ([]byte, error) {
	roomID, err := sqlite.NewQuery(db).Chat.GetRoomIdByGroupId(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return roomID, nil
}

func AddUserToGroupChat(ctx context.Context, db *sql.DB, tx *sql.Tx, groupID []byte, userID []byte) error {
	roomID, err := findGroupChatRoomID(ctx, db, groupID)
	if err != nil {
		return err
	}

	params := db_chat.AddRoomParticipantParams{
		RoomID: roomID,
		UserID: userID,
	}

	if tx != nil {
		err = sqlite.NewQuery(db).Chat.WithTx(tx).AddRoomParticipant(ctx, params)
	} else {
		err = sqlite.NewQuery(db).Chat.AddRoomParticipant(ctx, params)
	}

	if err != nil {
		if sqlite.CheckUniqueConstraint(err) {
			return nil
		}
		return err
	}

	return nil
}

func RemoveUserFromGroupChat(ctx context.Context, db *sql.DB, groupID []byte, userID []byte) error {
	roomID, err := findGroupChatRoomID(ctx, db, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return sqlite.NewQuery(db).Chat.RemoveRoomParticipant(ctx, db_chat.RemoveRoomParticipantParams{
		UserID: userID,
		RoomID: roomID,
	})
}
