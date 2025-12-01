-- name: CreateGroup :one
INSERT INTO groups (group_id, group_name, description, image, creator_id, created_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: CreateGroupMember :exec
INSERT INTO group_members (user_id, group_id, status, invited_by, created_at)
VALUES (?, ?, 'joined', NULL, ?);

-- name: GetGroupByName :one
SELECT group_name FROM groups WHERE group_name = ? LIMIT 1;

-- name: GetGroupsWithMemberCount :many
SELECT
    g.group_id,
    g.group_name,
    g.description,
    g.image,
    g.creator_id,
    g.created_at,
    COUNT(gm.user_id) as member_count
FROM groups g
LEFT JOIN group_members gm ON g.group_id = gm.group_id AND gm.status = 'joined'
GROUP BY g.group_id
ORDER BY g.created_at DESC;

-- name: IsGroupMember :one
SELECT COUNT(*) FROM group_members
WHERE user_id = ? AND group_id = ? AND status = 'joined';

-- name: GetGroupById :one
SELECT * FROM groups WHERE group_id = ?;

-- name: GetGroupMembers :many
SELECT 
    gm.user_id,
    gm.status,
    u.first_name,
    u.last_name,
    u.avatar
FROM group_members gm
JOIN users u ON gm.user_id = u.user_id
WHERE gm.group_id = ? AND gm.status = 'joined';

-- name: GetGroupEventsWithRSVPs :many
SELECT 
    ge.event_id,
    ge.title,
    ge.description,
    ge.event_timestamp,
    ge.created_at as event_created_at,
    gr.user_id as rsvp_user_id,
    gr.status as rsvp_status,
    gr.created_at as rsvp_created_at,
    u.first_name as rsvp_first_name,
    u.last_name as rsvp_last_name,
    u.avatar as rsvp_avatar
FROM group_events ge
LEFT JOIN group_rsvp gr ON ge.event_id = gr.event_id
LEFT JOIN users u ON gr.user_id = u.user_id
WHERE ge.group_id = ?
ORDER BY ge.event_timestamp DESC;

