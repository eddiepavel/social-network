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

-- name: GetGroupDetailsById :many
SELECT
    g.group_name as group_group_name,
    g.created_at as group_created_at,
    
    -- Group members
    gm.user_id as member_user_id,
    gm.status as member_status,
    gm.invited_by,
    gm.created_at as member_joined_at,
    u.email as member_email,
    u.first_name as member_first_name,
    u.last_name as member_last_name,
    u.avatar as member_avatar,
    u.nickname as member_nickname,
    u.about_me as member_about_me,
    
    -- Events
    ge.event_id,
    ge.title as event_title,
    ge.description as event_description,
    ge.event_timestamp,
    ge.created_at as event_created_at,
    
    -- RSVPs
    gr.user_id as rsvp_user_id,
    gr.status as rsvp_status,
    gr.created_at as rsvp_created_at,
    ru.first_name as rsvp_first_name,
    ru.last_name as rsvp_last_name,
    ru.avatar as rsvp_avatar,
    ru.nickname as rsvp_nickname

FROM groups g

LEFT JOIN group_members gm ON g.group_id = gm.group_id AND gm.status = 'joined'
LEFT JOIN users u ON gm.user_id = u.user_id

LEFT JOIN group_events ge ON g.group_id = ge.group_id

LEFT JOIN group_rsvp gr ON ge.event_id = gr.event_id
LEFT JOIN users ru ON gr.user_id = ru.user_id

WHERE g.group_id = ?;

