package models

import (
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
	GroupID     string `json:"group_id"`
	GroupName   string `json:"group_name"`
	Description string `json:"description"`
	Image       string `json:"image,omitempty"`
	ImageUrl    string `json:"image_url,omitempty"`
	CreatorID   string `json:"creator_id"`
	CreatedAt   string `json:"created_at"`
	MemberCount int64  `json:"member_count,omitempty"`
	IsOwner     bool   `json:"is_owner,omitempty"`
}

// GroupMemberResponse represents a group member returned to the client
type GroupMemberResponse struct {
	UserID    string  `json:"user_id"`
	GroupID   string  `json:"group_id,omitempty"`
	Status    string  `json:"status"`
	InvitedBy *string `json:"invited_by,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	// User details (populated via JOIN)
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
	CanRemove bool    `json:"can_remove_member,omitempty"`
}

// EventResponse represents a group event returned to the client
type EventResponse struct {
	EventID     string         `json:"event_id"`
	EventName   string         `json:"event_name"`
	Description string         `json:"description"`
	Timestamp   string         `json:"timestamp"`
	CreatedAt   time.Time      `json:"created_at"`
	RSVPs       []RSVPResponse `json:"rsvps"`
}

// RSVPResponse represents an event RSVP returned to the client
type RSVPResponse struct {
	UserID    string  `json:"user_id"`
	Status    string  `json:"status"` // 'going', 'not_going'
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
	CreatedAt string  `json:"created_at"`
}

// InviteUserRequest represents the request body for inviting a user to a group
type InviteUserRequest struct {
	UserID string `json:"user_id"`
}

// HandleJoinRequestRequest represents the request body for accepting/rejecting join requests
type HandleJoinRequestRequest struct {
	Action string `json:"action"` // 'accept' or 'reject'
}

// GroupDetailsResponse represents detailed group information
type GroupDetailsResponse struct {
	Group   GroupResponse         `json:"group"`
	Members []GroupMemberResponse `json:"members"`
	Events  []EventResponse       `json:"events"`
}

type InviteGroupRequest struct {
	Users []string `json:"users"`
}

type MemberShipRequest struct {
	Action string `json:"action"`
}
