package constants

// NotificationType represents the type of notification
// Go doesn't have native enums, so we use string constants with a custom type
type NotificationType string

const (
	// Follower notifications
	NotificationFollowRequest  NotificationType = "follow_request"
	NotificationFollowAccepted NotificationType = "follow_accepted"

	// Group notifications
	NotificationGroupInvitation   NotificationType = "group_invitation"
	NotificationGroupRequest      NotificationType = "group_request"
	NotificationGroupJoinApproved NotificationType = "group_join_approved"
	NotificationGroupJoinRejected NotificationType = "group_join_rejected"
	NotificationGroupEvent        NotificationType = "group_event"

	// Post notifications
	NotificationPostComment     NotificationType = "post_comment"
	NotificationCommentReply    NotificationType = "comment_reply"
	NotificationPostReaction    NotificationType = "post_reaction"
	NotificationCommentReaction NotificationType = "comment_reaction"

	// Message notifications
	NotificationMessage NotificationType = "message"
)

// String returns the string representation of NotificationType
func (n NotificationType) String() string {
	return string(n)
}

// IsValid checks if the notification type is valid
func (n NotificationType) IsValid() bool {
	switch n {
	case NotificationFollowRequest,
		NotificationFollowAccepted,
		NotificationGroupInvitation,
		NotificationGroupRequest,
		NotificationGroupJoinApproved,
		NotificationGroupJoinRejected,
		NotificationGroupEvent,
		NotificationPostComment,
		NotificationCommentReply,
		NotificationPostReaction,
		NotificationCommentReaction,
		NotificationMessage:
		return true
	}
	return false
}

// AllNotificationTypes returns all valid notification types
func AllNotificationTypes() []NotificationType {
	return []NotificationType{
		NotificationFollowRequest,
		NotificationFollowAccepted,
		NotificationGroupInvitation,
		NotificationGroupRequest,
		NotificationGroupJoinApproved,
		NotificationGroupJoinRejected,
		NotificationGroupEvent,
		NotificationPostComment,
		NotificationCommentReply,
		NotificationPostReaction,
		NotificationCommentReaction,
		NotificationMessage,
	}
}
