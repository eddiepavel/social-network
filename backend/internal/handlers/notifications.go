package handlers

import (
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_notifications "social-network/pkg/db/queries/notifications"
	"social-network/pkg/db/sqlite"
	"strconv"
)

// GetNotifications handles GET /api/notifications
// Returns the user's notifications with pagination
func GetNotifications(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Parse pagination parameters
		page := 1
		size := 20

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

		notifications, err := sqlite.NewQuery(app.DB).Notifications.GetNotifications(r.Context(),
			db_notifications.GetNotificationsParams{
				ReceiverID: currentUserID,
				Limit:      limit,
				Offset:     offset,
			})
		if err != nil {
			app.Logger.Error("failed to get notifications", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		unreadCount, err := sqlite.NewQuery(app.DB).Notifications.GetUnreadCount(r.Context(), currentUserID)
		if err != nil {
			app.Logger.Error("failed to get unread count", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		var notificationsList []models.NotificationResponse
		for _, n := range notifications {
			fromID, _ := helpers.GenerateFromBytes(n.FromID)

			notif := models.NotificationResponse{
				NotifID:   n.NotifID,
				Type:      n.Type,
				IsSeen:    n.IsSeen.Valid && n.IsSeen.Bool,
				FromID:    fromID,
				FromName:  n.FromFirstName.String + " " + n.FromLastName.String,
				CreatedAt: n.CreatedAt.Time,
			}

			if n.GroupID != nil {
				groupID, _ := helpers.GenerateFromBytes(n.GroupID)
				notif.GroupID = &groupID
				if n.GroupName.Valid {
					notif.GroupName = &n.GroupName.String
				}
			}

			if n.EventID != nil {
				eventID, _ := helpers.GenerateFromBytes(n.EventID)
				notif.EventID = &eventID
				if n.EventTitle.Valid {
					notif.EventTitle = &n.EventTitle.String
				}
			}

			notificationsList = append(notificationsList, notif)
		}

		response := models.NotificationsListResponse{
			Notifications: notificationsList,
			UnreadCount:   unreadCount,
		}

		utils.OK(w, response)
	}
}

// MarkNotificationAsRead handles POST /api/notifications/{id}/read
// Marks a single notification as read
func MarkNotificationAsRead(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		notifID := r.PathValue("id")
		if notifID == "" {
			utils.BadRequest(w, errors.New("notification ID required"))
			return
		}

		err := sqlite.NewQuery(app.DB).Notifications.MarkAsRead(r.Context(),
			db_notifications.MarkAsReadParams{
				NotifID:    notifID,
				ReceiverID: currentUserID,
			})
		if err != nil {
			app.Logger.Error("failed to mark notification as read", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]string{"message": "notification marked as read"})
	}
}

// MarkAllNotificationsAsRead handles POST /api/notifications/read-all
// Marks all notifications as read for the current user
func MarkAllNotificationsAsRead(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		err := sqlite.NewQuery(app.DB).Notifications.MarkAllAsRead(r.Context(), currentUserID)
		if err != nil {
			app.Logger.Error("failed to mark all notifications as read", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]string{"message": "all notifications marked as read"})
	}
}
