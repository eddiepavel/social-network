package models

import (
	"encoding/hex"
	"time"
)

// Group represents a group in the system
type Group struct {
	GroupID     []byte    `json:"group_id"`
	GroupName   string    `json:"group_name"`
	Description string    `json:"description"`
	Image       *string   `json:"image"`
	CreatorID   string    `json:"creator_id"` // TEXT in DB (inconsistency with schema)
	CreatedAt   time.Time `json:"created_at"`
}

// GroupMember represents a group membership
type GroupMember struct {
	UserID    []byte    `json:"user_id"`
	GroupID   []byte    `json:"group_id"`
	Status    string    `json:"status"` // 'joined', 'requested', 'rejected'
	InvitedBy *[]byte   `json:"invited_by"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateGroupRequest represents the request body for creating a group
type CreateGroupRequest struct {
	GroupName   string `json:"group_name"`
	Description string `json:"description"`
	Image       string `json:"image,omitempty"`
}

// GroupResponse represents a group returned to the client
type GroupResponse struct {
	GroupID     string  `json:"group_id"`
	GroupName   string  `json:"group_name"`
	Description string  `json:"description"`
	Image       *string `json:"image"`
	CreatorID   string  `json:"creator_id"`
	CreatedAt   string  `json:"created_at"`
	MemberCount int64   `json:"member_count"`
}

// GroupMemberResponse represents a group member returned to the client
type GroupMemberResponse struct {
	UserID    string  `json:"user_id"`
	GroupID   string  `json:"group_id"`
	Status    string  `json:"status"`
	InvitedBy *string `json:"invited_by,omitempty"`
	CreatedAt string  `json:"created_at"`
	// User details (populated via JOIN)
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
}

// InviteUserRequest represents the request body for inviting a user to a group
type InviteUserRequest struct {
	UserID string `json:"user_id"`
}

// HandleJoinRequestRequest represents the request body for accepting/rejecting join requests
type HandleJoinRequestRequest struct {
	Action string `json:"action"` // 'accept' or 'reject'
}

type Events struct {
	EventName string `json:"event_name"`
}

// GroupDetailsResponse represents detailed group information
type GroupDetailsResponse struct {
	GroupResponse `json:"group"`
	Members       []GroupMemberResponse `json:"members"`
	Events        []Events
}

// Helper function to convert binary UUID to hex string
func UUIDBytesToString(b []byte) string {
	if b == nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// Helper function to convert hex string to binary UUID
func UUIDStringToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
