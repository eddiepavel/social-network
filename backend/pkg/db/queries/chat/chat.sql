-- name: CreateMessage :exec
INSERT INTO chat_messages (message_id, content, sender_id, target_id) VALUES (?, ?, ?, ?);

-- name: CreateRoom :exec
INSERT INTO chat_rooms (room_id, name, group_id) VALUES (?, ?, ?);

-- name: AddRoomParticipant :exec
INSERT INTO chat_participants (room_id, user_id) VALUES (?, ?);

-- name: GetRoomIdByGroupId :one
SELECT cr.room_id as room_id FROM chat_rooms cr WHERE cr.group_id = ?;

-- name: GetOtherRoomParticipants :many
SELECT u.user_id, u.first_name, u.last_name, u.avatar FROM users u JOIN chat_participants cp ON cp.user_id = u.user_id WHERE cp.room_id = ? AND cp.user_id != ?;

-- name: CheckIfRoomIsGroup :one
SELECT cr.group_id FROM chat_rooms cr JOIN groups g on g.group_id = cr.group_id WHERE room_id = ?;

-- name: GetUserChatList :many
SELECT
    cr.room_id,
    cr.name AS room_name,
    cr.group_id,
    cr.created_at AS room_created_at,

    lm.message_id AS last_message_id,
    lm.content AS last_message_content,
    lm.created_at AS last_message_time,
    lm.sender_id AS last_message_sender_id,

    CAST(COALESCE(
            (SELECT COUNT(*)
             FROM chat_messages cm
             WHERE cm.target_id = cr.room_id
               AND cm.sender_id != ?
           AND cm.created_at >= COALESCE(cp.last_read_at, cp.joined_at, '1970-01-01')
        ), 0
    ) AS INTEGER) AS unread_count,

    cp.last_read_at,
    cp.joined_at

FROM chat_rooms cr

         INNER JOIN chat_participants cp
                    ON cr.room_id = cp.room_id
                        AND cp.user_id = ?

         INNER JOIN (
    SELECT cm1.*
    FROM chat_messages cm1
             INNER JOIN (
        SELECT target_id, MAX(created_at) AS max_time
        FROM chat_messages
        GROUP BY target_id
    ) cm2 ON cm1.target_id = cm2.target_id
        AND cm1.created_at = cm2.max_time
) lm ON cr.room_id = lm.target_id

ORDER BY lm.created_at DESC;

-- name: GetRoomMessages :many
SELECT
    cm.message_id,
    cm.content,
    cm.sender_id,
    cm.target_id,
    cm.created_at,
    u.first_name AS sender_first_name,
    u.last_name AS sender_last_name,
    u.avatar AS sender_avatar
FROM chat_messages cm
JOIN users u ON cm.sender_id = u.user_id
WHERE cm.target_id = ?
  AND (? IS NULL OR cm.created_at <= ?)
  AND (? IS NULL OR cm.message_id != ?)
ORDER BY cm.created_at DESC, cm.message_id DESC
    LIMIT ?;

-- name: GetRoomMessagesCount :one
SELECT COUNT(*) FROM chat_messages WHERE target_id = ?;

-- name: MarkRoomMessagesAsRead :exec
UPDATE chat_participants SET last_read_at = CURRENT_TIMESTAMP WHERE user_id = ? AND room_id = ?;

-- name: RemoveRoomParticipant :exec
DELETE FROM chat_participants WHERE user_id = ? AND room_id = ?;

-- name: CheckUserIsParticipant :one
SELECT COUNT(*) FROM chat_participants WHERE user_id = ? AND room_id = ?;

-- name: FindRoomBetweenUsers :one
SELECT cr.room_id
FROM chat_rooms cr
         INNER JOIN chat_participants cp1
                    ON cr.room_id = cp1.room_id
                        AND cp1.user_id = ?
         INNER JOIN chat_participants cp2
                    ON cr.room_id = cp2.room_id
                        AND cp2.user_id = ?
WHERE cr.group_id != null
  AND cp1.user_id != cp2.user_id  -- Ensure different users
LIMIT 1;

-- name: GetRoomParticipants :many
SELECT user_id FROM chat_participants WHERE room_id = ?;

-- name: GetUserBasicInfo :one
SELECT first_name, last_name, avatar FROM users WHERE user_id = ?;

-- name: UpdateRoomName :exec
UPDATE chat_rooms SET name = ? WHERE room_id = ?;
