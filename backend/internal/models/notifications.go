package models

import "time"

// NotificationResponse represents a notification returned to the client
type NotificationResponse struct {
	NotifID     string    `json:"notif_id"`
	Type        string    `json:"type"` // follow_request, group_invitation, group_request, group_event, message
	IsSeen      bool      `json:"is_seen"`
	FromID      string    `json:"from_id"`
	FromName    string    `json:"from_name,omitempty"`
	GroupID     *string   `json:"group_id,omitempty"`
	GroupName   *string   `json:"group_name,omitempty"`
	EventID     *string   `json:"event_id,omitempty"`
	EventTitle  *string   `json:"event_title,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// NotificationsListResponse represents the list of notifications with unread count
type NotificationsListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
}
