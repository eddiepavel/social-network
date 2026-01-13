-- name: CreateGroup :one
INSERT INTO groups (group_id, group_name, description, image, creator_id, created_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: CreateGroupMember :exec
INSERT INTO group_members (user_id, group_id, status, invited_by, created_at)
VALUES (?, ?, ?, NULL, ?);

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
    i.image_id as group_image,
    i.image_path as group_image_path,
    i.file_name as group_image_file_name,
    COUNT(gm.user_id) as member_count
FROM groups g
LEFT JOIN group_members gm ON g.group_id = gm.group_id AND gm.status = 'joined'
LEFT JOIN images i ON g.image = i.image_id
GROUP BY g.group_id
ORDER BY g.created_at DESC;

-- name: IsGroupMember :one
SELECT * FROM group_members
WHERE user_id = ? AND group_id = ? LIMIT 1;

-- name: RemoveUserFromGroup :exec
DELETE FROM group_members WHERE user_id = ? AND group_id = ?;

-- name: GetGroupById :one
SELECT
    g.group_id,
    g.group_name,
    g.description,
    g.image,
    g.creator_id,
    g.created_at,
    i.image_id as group_image_id,
    i.image_path as group_image_path,
    i.file_name as group_image_file_name
FROM groups g
JOIN images i ON g.image = i.image_id
WHERE g.group_id = ?;

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

-- name: GetGroupMembersWithRequests :many
SELECT * FROM group_members WHERE group_id = ?;

-- name: UpdateGroupMemberStatus :exec
UPDATE group_members SET status = ? WHERE user_id = ?; 

-- name: InviteGroupMembers :exec
INSERT INTO group_members (user_id, group_id, status, invited_by) VALUES (?, ?, ?, ?);

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

-- name: GetGroupJoinRequests :many
SELECT * FROM group_members WHERE group_id = ? AND status = 'requested';

-- name: DeleteDbGroup :exec
DELETE FROM groups WHERE group_id = ?;

-- name: UpdateDbGroup :one
UPDATE groups SET group_name = ?, description = ?, image = ? 
WHERE group_id = ? AND creator_id = ? RETURNING *;

