package helpers

import (
	"context"
	"social-network/internal/models"
	"social-network/pkg/db/sqlite"
)

func CreateGroupDetailResponse(groupId []byte, t *sqlite.Transactions) (models.GroupDetailsResponse, error) {

	var group models.GroupDetailsResponse
	ctx := context.Background()

	// Get group
	getGroup, err := t.Groups.GetGroupById(ctx, groupId)
	if err != nil {
		return group, err
	}
	groupUUID, _ := GenerateFromBytes(getGroup.GroupID)

	group.Group = models.GroupResponse{
		GroupID:     groupUUID,
		GroupName:   getGroup.GroupName,
		Description: getGroup.Description,
		CreatedAt:   getGroup.CreatedAt.Local().String(),
	}
	if getGroup.Image.Valid {
		group.Group.Image = &getGroup.Image.String
	}

	// Get members
	getMembers, _ := t.Groups.GetGroupMembers(ctx, groupId)
	for _, member := range getMembers {
		memberUUID, _ := GenerateFromBytes(member.UserID)
		memberResp := models.GroupMemberResponse{
			UserID:    memberUUID,
			Status:    member.Status,
			FirstName: &member.FirstName,
			LastName:  &member.LastName,
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
