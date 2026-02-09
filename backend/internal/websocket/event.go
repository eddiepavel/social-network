package websocket

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function that handles incoming events from clients
type EventHandler func(event Event, c *Client, db *sql.DB) error

const (
	// Server -> Client events (outgoing)
	EventNotification = "notification"
	EventError        = "error"

	// Client -> Server events (incoming)
	EventPrivateMessage  = "private_message"
	EventTypingIndicator = "typing_indicator"
)

// NotificationEvent represents a notification pushed to the client
type NotificationEvent struct {
	NotifID    string    `json:"notif_id"`
	ReceiverID string    `json:"receiver_id"`
	Type       string    `json:"type"`
	FromID     string    `json:"from_id"`
	GroupID    string    `json:"group_id,omitempty"`
	EventID    string    `json:"event_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	IsSeen     bool      `json:"is_seen"`
}

// ErrorEvent represents an error message sent to the client
type ErrorEvent struct {
	Message string `json:"message"`
}

// PrivateMessageEvent represents a private message between users
type PrivateMessageEvent struct {
	From    string    `json:"from_user"`
	To      string    `json:"to_user"`
	Message string    `json:"message"`
	Sent    time.Time `json:"sent"`
}

// TypingIndicatorEvent represents a typing indicator
type TypingIndicatorEvent struct {
	From     string `json:"from_user"`
	To       string `json:"to_user"`
	IsTyping bool   `json:"is_typing"`
}
