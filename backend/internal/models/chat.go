package models

import (
	"time"
)

type ChatList struct {
	RoomID              string    `json:"room_id"`
	RoomName            *string   `json:"room_name,omitempty"`
	IsGroup             bool      `json:"is_group"`
	LastMessageID       string    `json:"last_message_id,omitempty"`
	LastMessageContent  string    `json:"last_message_content,omitempty"`
	LastMessageTime     time.Time `json:"last_message_time,omitempty"`
	LastMessageSenderID string    `json:"last_message_sender_id,omitempty"`
	UnreadCount         int       `json:"unread_count"`
}

type ChatMessages struct {
	MessageID string    `json:"message_id"`
	Content   string    `json:"content"`
	SenderID  string    `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatMessageResponse struct {
	Messages   []ChatMessages `json:"messages"`
	HasMore    bool           `json:"has_more"`
	NextCursor time.Time      `json:"next_cursor"`
}

type CreateMessageRequest struct {
	Content string `json:"content"`
}

type FirstCreateMessageRequest struct {
	TargetID string `json:"target_id"`
	Content  string `json:"content"`
}
