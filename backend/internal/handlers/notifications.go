package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	"social-network/pkg/db/sqlite"
	"time"
)

// GetNotifications handles GET /api/notifications/
// Returns all notifications for the authenticated user
func GetNotifications(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Fetch all notifications for this user
		notifications, err := sqlite.NewQuery(app.DB).Notifications.GetNotificationsByReceiverId(r.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.OK(w, []models.NotificationResponse{})
				return
			}
			app.Logger.Error("failed to fetch notifications", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Convert to response format
		var response []models.NotificationResponse
		for _, notif := range notifications {
			receiverUUID, _ := helpers.GenerateFromBytes(notif.ReceiverID)
			fromUUID, _ := helpers.GenerateFromBytes(notif.FromID)

			response = append(response, models.NotificationResponse{
				NotifID:    notif.NotifID,
				ReceiverID: receiverUUID,
				Type:       notif.Type,
				IsSeen:     notif.IsSeen.Bool,
				FromID:     fromUUID,
				GroupID: func() *string {
					if len(notif.GroupID) > 0 {
						groupUUID, _ := helpers.GenerateFromBytes(notif.GroupID)
						return &groupUUID
					}
					return nil
				}(),
				EventID: func() *string {
					if len(notif.EventID) > 0 {
						eventUUID, _ := helpers.GenerateFromBytes(notif.EventID)
						return &eventUUID
					}
					return nil
				}(),
				CreatedAt: func() string {
					if notif.CreatedAt.Valid {
						return notif.CreatedAt.Time.Format(time.RFC3339)
					}
					return ""
				}(),
			})
		}

		utils.OK(w, response)
	}
}

// GetUnseenNotifications handles GET /api/notifications/unseen
// Returns all unseen notifications for the authenticated user
func GetUnseenNotifications(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Fetch unseen notifications for this user
		notifications, err := sqlite.NewQuery(app.DB).Notifications.GetUnseenNotificationsByReceiverId(r.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.OK(w, []models.NotificationResponse{})
				return
			}
			app.Logger.Error("failed to fetch unseen notifications", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Convert to response format
		var response []models.NotificationResponse
		for _, notif := range notifications {
			receiverUUID, _ := helpers.GenerateFromBytes(notif.ReceiverID)
			fromUUID, _ := helpers.GenerateFromBytes(notif.FromID)

			response = append(response, models.NotificationResponse{
				NotifID:    notif.NotifID,
				ReceiverID: receiverUUID,
				Type:       notif.Type,
				IsSeen:     notif.IsSeen.Bool,
				FromID:     fromUUID,
				GroupID: func() *string {
					if len(notif.GroupID) > 0 {
						groupUUID, _ := helpers.GenerateFromBytes(notif.GroupID)
						return &groupUUID
					}
					return nil
				}(),
				EventID: func() *string {
					if len(notif.EventID) > 0 {
						eventUUID, _ := helpers.GenerateFromBytes(notif.EventID)
						return &eventUUID
					}
					return nil
				}(),
				CreatedAt: func() string {
					if notif.CreatedAt.Valid {
						return notif.CreatedAt.Time.Format(time.RFC3339)
					}
					return ""
				}(),
			})
		}

		utils.OK(w, response)
	}
}

// GetNotificationsWithUserDetails handles GET /api/notifications/details
// Returns notifications with sender user details (JOIN with users table)
func GetNotificationsWithUserDetails(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Fetch notifications with user details
		notifications, err := sqlite.NewQuery(app.DB).Notifications.GetNotificationWithUserDetails(r.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.OK(w, []models.NotificationWithUserDetailsResponse{})
				return
			}
			app.Logger.Error("failed to fetch notifications with user details", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Convert to response format
		var response []models.NotificationWithUserDetailsResponse
		for _, notif := range notifications {
			receiverUUID, _ := helpers.GenerateFromBytes(notif.ReceiverID)
			fromUUID, _ := helpers.GenerateFromBytes(notif.FromID)

			response = append(response, models.NotificationWithUserDetailsResponse{
				NotifID:    notif.NotifID,
				ReceiverID: receiverUUID,
				Type:       notif.Type,
				IsSeen:     notif.IsSeen.Bool,
				FromID:     fromUUID,
				GroupID: func() *string {
					if len(notif.GroupID) > 0 {
						groupUUID, _ := helpers.GenerateFromBytes(notif.GroupID)
						return &groupUUID
					}
					return nil
				}(),
				EventID: func() *string {
					if len(notif.EventID) > 0 {
						eventUUID, _ := helpers.GenerateFromBytes(notif.EventID)
						return &eventUUID
					}
					return nil
				}(),
				CreatedAt: func() string {
					if notif.CreatedAt.Valid {
						return notif.CreatedAt.Time.Format(time.RFC3339)
					}
					return ""
				}(),
				FromName: notif.FromFirstName + " " + notif.FromLastName,
				FromAvatar: func() *string {
					if notif.FromAvatar.Valid && notif.FromAvatar.String != "" {
						img := app.File.GenerateSignImage(notif.FromAvatar.String, userID, time.Now().Add(15*time.Minute))
						return &img
					}
					return nil
				}(),
				FromNickname: func() *string {
					if notif.FromNickname.Valid {
						return &notif.FromNickname.String
					}
					return nil
				}(),
			})
		}

		utils.OK(w, response)
	}
}

// MarkNotificationAsSeen handles PUT /api/notifications/{notificationId}/seen
// Marks a specific notification as seen
func MarkNotificationAsSeen(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Get notification ID from path
		notifID := r.PathValue("notificationId")
		if notifID == "" {
			utils.BadRequest(w, errors.New("notification id required"))
			return
		}

		// Check if notification exists and belongs to user
		notif, err := sqlite.NewQuery(app.DB).Notifications.GetNotificationById(r.Context(), notifID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
			app.Logger.Error("failed to fetch notification", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Verify notification belongs to current user
		notifReceiverUUID, _ := helpers.GenerateFromBytes(notif.ReceiverID)
		currentUserUUID, _ := helpers.GenerateFromBytes(userID)
		if notifReceiverUUID != currentUserUUID {
			utils.Forbidden(w)
			return
		}

		// Mark as seen
		err = sqlite.NewQuery(app.DB).Notifications.MarkNotificationAsSeen(r.Context(), notifID)
		if err != nil {
			app.Logger.Error("failed to mark notification as seen", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]string{"message": "notification marked as seen"})
	}
}

// MarkAllNotificationsAsSeen handles PUT /api/notifications/seen/all
// Marks all notifications as seen for the authenticated user
func MarkAllNotificationsAsSeen(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Mark all notifications as seen
		err := sqlite.NewQuery(app.DB).Notifications.MarkAllNotificationsAsSeenForUser(r.Context(), userID)
		if err != nil {
			app.Logger.Error("failed to mark all notifications as seen", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]string{"message": "all notifications marked as seen"})
	}
}

// DeleteNotification handles DELETE /api/notifications/{notificationId}
// Deletes a specific notification
func DeleteNotification(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Get notification ID from path
		notifID := r.PathValue("notificationId")
		if notifID == "" {
			utils.BadRequest(w, errors.New("notification id required"))
			return
		}

		// Check if notification exists and belongs to user
		notif, err := sqlite.NewQuery(app.DB).Notifications.GetNotificationById(r.Context(), notifID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.NotFound(w)
				return
			}
			app.Logger.Error("failed to fetch notification", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		// Verify notification belongs to current user
		notifReceiverUUID, _ := helpers.GenerateFromBytes(notif.ReceiverID)
		currentUserUUID, _ := helpers.GenerateFromBytes(userID)
		if notifReceiverUUID != currentUserUUID {
			utils.Forbidden(w)
			return
		}

		// Delete notification
		err = sqlite.NewQuery(app.DB).Notifications.DeleteNotification(r.Context(), notifID)
		if err != nil {
			app.Logger.Error("failed to delete notification", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, map[string]string{"message": "notification deleted"})
	}
}

// GetUnseenNotificationCount handles GET /api/notifications/unseen/count
// Returns the count of unseen notifications for the authenticated user
func GetUnseenNotificationCount(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		// Get count
		count, err := sqlite.NewQuery(app.DB).Notifications.CountUnseenNotifications(r.Context(), userID)
		if err != nil {
			app.Logger.Error("failed to count unseen notifications", "err", err)
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		utils.OK(w, models.UnseenCountResponse{Count: count})
	}
}
