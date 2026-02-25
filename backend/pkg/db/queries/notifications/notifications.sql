-- name: CreateNotification :one
INSERT INTO notifications (notif_id, receiver_id, type, from_id, group_id, event_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetNotificationById :one
SELECT * FROM notifications WHERE notif_id = ? LIMIT 1;

-- name: GetNotificationsByReceiverId :many
SELECT * FROM notifications
WHERE receiver_id = ?
ORDER BY created_at DESC;

-- name: GetUnseenNotificationsByReceiverId :many
SELECT * FROM notifications
WHERE receiver_id = ? AND is_seen = 0
ORDER BY created_at DESC;

-- name: GetNotificationsByType :many
SELECT * FROM notifications
WHERE receiver_id = ? AND type = ?
ORDER BY created_at DESC;

-- name: MarkNotificationAsSeen :exec
UPDATE notifications
SET is_seen = 1
WHERE notif_id = ?;

-- name: MarkAllNotificationsAsSeenForUser :exec
UPDATE notifications
SET is_seen = 1
WHERE receiver_id = ? AND is_seen = 0;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE notif_id = ?;

-- name: DeleteNotificationsByReceiverId :exec
DELETE FROM notifications WHERE receiver_id = ?;

-- name: CountUnseenNotifications :one
SELECT COUNT(*) FROM notifications
WHERE receiver_id = ? AND is_seen = 0;

-- name: GetNotificationWithUserDetails :many
SELECT
    n.notif_id,
    n.receiver_id,
    n.type,
    n.is_seen,
    n.from_id,
    n.group_id,
    n.event_id,
    n.created_at,
    u.first_name as from_first_name,
    u.last_name as from_last_name,
    u.avatar as from_avatar,
    u.nickname as from_nickname
FROM notifications n
JOIN users u ON n.from_id = u.user_id
WHERE n.receiver_id = ?
ORDER BY n.created_at DESC;

-- name: GetUnseenNotificationsWithUserDetails :many
SELECT
    n.notif_id,
    n.receiver_id,
    n.type,
    n.is_seen,
    n.from_id,
    n.group_id,
    n.event_id,
    n.created_at,
    u.first_name as from_first_name,
    u.last_name as from_last_name,
    u.avatar as from_avatar,
    u.nickname as from_nickname
FROM notifications n
JOIN users u ON n.from_id = u.user_id
WHERE n.receiver_id = ? AND n.is_seen = 0
ORDER BY n.created_at DESC;

-- name: GetLastMessageNotification :one
SELECT created_at FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = 'message'
ORDER BY created_at DESC
LIMIT 1;
-- name: GetLastFollowRequestNotification :one
SELECT created_at FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = 'follow_request'
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLastFollowAcceptedNotification :one
SELECT created_at FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = 'follow_accepted'
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLastGroupRequestNotification :one
SELECT created_at FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = 'group_request'
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLastGroupJoinApprovedNotification :one
SELECT created_at FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = 'group_join_approved'
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteFollowRequestNotification :exec
DELETE FROM notifications 
WHERE receiver_id = ? AND from_id = ? AND type = 'follow_request';

-- name: DeleteGroupRequestNotification :exec
DELETE FROM notifications
WHERE receiver_id = ? AND from_id = ? AND group_id = ? AND type = 'group_request';

-- name: GetLastPostReactionNotification :one
SELECT created_at FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = 'post_reaction'
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLastCommentReactionNotification :one
SELECT created_at FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = 'comment_reaction'
ORDER BY created_at DESC
LIMIT 1;

-- name: DeletePostReactionNotification :exec
DELETE FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = 'post_reaction';

-- name: DeleteCommentReactionNotification :exec
DELETE FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = 'comment_reaction';

-- name: DeleteCommentNotificationByReceiverAndType :exec
DELETE FROM notifications
WHERE receiver_id = ? AND from_id = ? AND type = ?;
