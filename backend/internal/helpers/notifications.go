package helpers

import (
	"context"
	"database/sql"
	db_notifications "social-network/pkg/db/queries/notifications"
	"social-network/pkg/db/sqlite"

	"github.com/google/uuid"
)

// CreateNotification creates a notification in the database
// notifType: follow_request, group_invitation, group_request, group_event, message
func CreateNotification(
	db *sql.DB,
	ctx context.Context,
	receiverID []byte,
	notifType string,
	fromID []byte,
	groupID []byte,
	eventID []byte,
) error {
	notifID := uuid.New().String()

	return sqlite.NewQuery(db).Notifications.CreateNotification(ctx,
		db_notifications.CreateNotificationParams{
			NotifID:    notifID,
			ReceiverID: receiverID,
			Type:       notifType,
			FromID:     fromID,
			GroupID:    groupID,
			EventID:    eventID,
		})
}
