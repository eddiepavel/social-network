package models

import (
	"time"
)

type ChatList struct {
	RoomID             string       `json:"room_id"`
	RoomName           *string      `json:"room_name,omitempty"`
	CanEditName        bool         `json:"can_edit_room_name"`
	GroupID            *string      `json:"group_id,omitempty"`
	OtherUser          UserResponse `json:"other_user,omitempty"`
	LastMessageID      string       `json:"last_message_id,omitempty"`
	LastMessageContent string       `json:"last_message_content,omitempty"`
	LastMessageTime    time.Time    `json:"last_message_time,omitempty"`
	LastMessageSender  UserResponse `json:"last_message_sender,omitempty"`
	UnreadCount        int          `json:"unread_count"`
}

type ChatMessages struct {
	MessageID       string    `json:"message_id"`
	Content         string    `json:"content"`
	SenderID        string    `json:"sender_id"`
	CreatedAt       time.Time `json:"created_at"`
	SenderFirstName string    `json:"sender_first_name,omitempty"`
	SenderLastName  string    `json:"sender_last_name,omitempty"`
	SenderAvatar    string    `json:"sender_avatar,omitempty"`
}

type ChatMessageResponse struct {
	Messages   []ChatMessages   `json:"messages"`
	HasMore    bool             `json:"has_more"`
	NextCursor CursorPagination `json:"next_cursor"`
}

type CursorPagination struct {
	CursorTimestamp time.Time `json:"cursor_timestamp"`
	CursorID        string    `json:"cursor_id"`
}

type CreateMessageRequest struct {
	Content string `json:"content"`
}

type FirstCreateMessageRequest struct {
	TargetID string `json:"target_id"`
	Content  string `json:"content"`
}

type EditRoomNameRequest struct {
	RoomName string `json:"room_name"`
}
