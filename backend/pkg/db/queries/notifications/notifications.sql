-- name: GetNotifications :many
SELECT n.*,
       u.first_name as from_first_name,
       u.last_name as from_last_name,
       g.group_name,
       e.title as event_title
FROM notifications n
LEFT JOIN users u ON n.from_id = u.user_id
LEFT JOIN groups g ON n.group_id = g.group_id
LEFT JOIN group_events e ON n.event_id = e.event_id
WHERE n.receiver_id = ?
ORDER BY n.created_at DESC
LIMIT ? OFFSET ?;

-- name: GetUnreadCount :one
SELECT COUNT(*) FROM notifications WHERE receiver_id = ? AND is_seen = 0;

-- name: MarkAsRead :exec
UPDATE notifications SET is_seen = 1 WHERE notif_id = ? AND receiver_id = ?;

-- name: MarkAllAsRead :exec
UPDATE notifications SET is_seen = 1 WHERE receiver_id = ?;

-- name: CreateNotification :exec
INSERT INTO notifications (notif_id, receiver_id, type, from_id, group_id, event_id)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE notif_id = ? AND receiver_id = ?;
