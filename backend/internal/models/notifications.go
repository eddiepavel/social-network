package models

import "time"

// NotificationResponse represents a notification returned to the client
type NotificationResponse struct {
	NotifID    string  `json:"notif_id"`
	ReceiverID string  `json:"receiver_id"`
	Type       string  `json:"type"`
	IsSeen     bool    `json:"is_seen"`
	FromID     string  `json:"from_id"`
	GroupID    *string `json:"group_id,omitempty"`
	EventID    *string `json:"event_id,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

// NotificationWithUserDetailsResponse represents a notification with sender details
type NotificationWithUserDetailsResponse struct {
	NotifID      string  `json:"notif_id"`
	ReceiverID   string  `json:"receiver_id"`
	Type         string  `json:"type"`
	IsSeen       bool    `json:"is_seen"`
	FromID       string  `json:"from_id"`
	GroupID      *string `json:"group_id,omitempty"`
	EventID      *string `json:"event_id,omitempty"`
	CreatedAt    string  `json:"created_at"`
	FromName     string  `json:"from_name"`
	FromAvatar   *string `json:"from_avatar,omitempty"`
	FromNickname *string `json:"from_nickname,omitempty"`
}

// CreateNotificationRequest represents the request body for creating a notification
type CreateNotificationRequest struct {
	ReceiverID string  `json:"receiver_id"`
	Type       string  `json:"type"`
	FromID     string  `json:"from_id"`
	GroupID    *string `json:"group_id,omitempty"`
	EventID    *string `json:"event_id,omitempty"`
}

// UnseenCountResponse represents the count of unseen notifications
type UnseenCountResponse struct {
	Count int64 `json:"count"`
}

// Notification represents a notification in the database
type Notification struct {
	NotifID    string     `json:"notif_id"`
	ReceiverID []byte     `json:"receiver_id"`
	Type       string     `json:"type"`
	IsSeen     bool       `json:"is_seen"`
	FromID     []byte     `json:"from_id"`
	GroupID    []byte     `json:"group_id,omitempty"`
	EventID    []byte     `json:"event_id,omitempty"`
	CreatedAt  *time.Time `json:"created_at"`
}
