package helpers

import (
	"context"
	"database/sql"
	"errors"
	"social-network/app"
	"social-network/internal/constants"
	db_notifications "social-network/pkg/db/queries/notifications"
	"social-network/pkg/db/sqlite"
	"time"

	"github.com/google/uuid"
)

// CreateNotification creates a notification
func CreateNotification(app *app.App, receiverID []byte, notifType constants.NotificationType, fromID []byte, groupID []byte, eventID []byte) error {
	// Validate notification type
	if !notifType.IsValid() {
		app.Logger.Error("invalid notification type", "type", notifType)
		return errors.New("invalid notification type")
	}

	// Generate notification ID
	notifUUID := uuid.New().String()

	// Prepare parameters
	params := db_notifications.CreateNotificationParams{
		NotifID:    notifUUID,
		ReceiverID: receiverID,
		Type:       notifType.String(),
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
