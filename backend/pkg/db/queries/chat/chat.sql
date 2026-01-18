-- name: CreateMessage :exec
INSERT INTO chat_messages (message_id, content, sender_id, target_id) VALUES (?, ?, ?, ?);

-- name: CreateRoom :exec
INSERT INTO chat_rooms (room_id, name, is_group) VALUES (?, ?, ?);

-- name: AddRoomParticipant :exec
INSERT INTO chat_participants (room_id, user_id) VALUES (?, ?);

-- name: GetUserChatList :many
SELECT
    cr.room_id,
    cr.name AS room_name,
    cr.is_group,
    cr.created_at AS room_created_at,

    lm.message_id AS last_message_id,
    lm.content AS last_message_content,
    lm.created_at AS last_message_time,
    lm.sender_id AS last_message_sender_id,

    CASE WHEN cr.is_group = 0 THEN other_user.user_id ELSE NULL END AS other_user_id,

    CAST(COALESCE(
            (SELECT COUNT(*)
             FROM chat_messages cm
             WHERE cm.target_id = cr.room_id
               AND cm.sender_id != ?
           AND cm.created_at > COALESCE(cp.last_read_at, cp.joined_at, '1970-01-01')
        ), 0
    ) AS INTEGER) AS unread_count,

    cp.last_read_at,
    cp.joined_at

FROM chat_rooms cr

         INNER JOIN chat_participants cp
                    ON cr.room_id = cp.room_id
                        AND cp.user_id = ?

         LEFT JOIN (
    SELECT cm1.*
    FROM chat_messages cm1
             INNER JOIN (
        SELECT target_id, MAX(created_at) AS max_time
        FROM chat_messages
        GROUP BY target_id
    ) cm2 ON cm1.target_id = cm2.target_id
        AND cm1.created_at = cm2.max_time
) lm ON cr.room_id = lm.target_id

         LEFT JOIN chat_participants other_cp
                   ON cr.room_id = other_cp.room_id
                       AND other_cp.user_id != ?
    AND cr.is_group = 0
LEFT JOIN users other_user ON other_cp.user_id = other_user.user_id

ORDER BY COALESCE(lm.created_at, cr.created_at) DESC;

-- name: GetRoomMessages :many
SELECT
    cm.message_id,
    cm.content,
    cm.sender_id,
    cm.target_id,
    cm.created_at
FROM chat_messages cm
WHERE cm.target_id = ?
  AND (? IS NULL OR cm.created_at < ?)  -- Cursor pagination
ORDER BY cm.created_at DESC
    LIMIT ?;

-- name: GetRoomMessagesCount :one
SELECT COUNT(*) FROM chat_messages WHERE target_id = ?;

-- name: MarkRoomMessagesAsRead :exec
UPDATE chat_participants SET last_read_at = ? WHERE user_id = ? AND room_id = ?;

-- name: RemoveRoomParticipant :exec
DELETE FROM chat_participants WHERE user_id = ? AND room_id = ?;

-- name: CheckUserIsParticipant :one
SELECT COUNT(*) FROM chat_participants WHERE user_id = ? AND room_id = ?;


