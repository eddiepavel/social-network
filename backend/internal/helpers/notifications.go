package helpers

import (
	"context"
	"database/sql"
	"social-network/app"
	db_notifications "social-network/pkg/db/queries/notifications"
	"social-network/pkg/db/sqlite"
	"time"

	"github.com/google/uuid"
)

// CreateNotification creates a notification
func CreateNotification(app *app.App, receiverID []byte, notifType string, fromID []byte, groupID []byte, eventID []byte) error {
	// Generate notification ID
	notifUUID := uuid.New().String()

	// Prepare parameters
	params := db_notifications.CreateNotificationParams{
		NotifID:    notifUUID,
		ReceiverID: receiverID,
		Type:       notifType,
		FromID:     fromID,
		GroupID:    groupID,
		EventID:    eventID,
		CreatedAt:  sql.NullTime{Valid: true, Time: time.Now()},
	}

	// Create notification using background context
	_, err := sqlite.NewQuery(app.DB).Notifications.CreateNotification(context.Background(), params)
	if err != nil {
		app.Logger.Error("failed to create notification", "err", err, "type", notifType)
		return err
	}

	app.Logger.Info("notification created", "type", notifType)
	return nil
}
