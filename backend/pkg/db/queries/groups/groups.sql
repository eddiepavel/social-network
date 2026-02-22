-- name: CreateGroup :one
INSERT INTO groups (group_id, group_name, description, image, creator_id, created_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: CreateGroupMember :exec
INSERT INTO group_members (user_id, group_id, status, invited_by, created_at)
VALUES (?, ?, ?, NULL, ?);

-- name: GetGroupByName :one
SELECT * FROM groups WHERE group_name = ? LIMIT 1;

-- name: GetGroupEventRSVPs :many
SELECT * FROM group_rsvp WHERE event_id = ?;

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
LEFT JOIN images i ON g.image = i.image_id
WHERE g.group_id = ?;

-- name: CheckIsCreator :one
SELECT EXISTS(
    SELECT 1 FROM groups WHERE creator_id = ? AND group_id = ?
) AS is_creator;

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
    ge.creator_id,
    ge.created_at as event_created_at,
    gr.user_id as rsvp_user_id,
    gr.status as rsvp_status,
    gr.created_at as rsvp_created_at,
    cu.first_name as creator_first_name,
    cu.last_name as creator_last_name,
    cu.avatar as creator_avatar,
    u.first_name as rsvp_first_name,
    u.last_name as rsvp_last_name,
    u.avatar as rsvp_avatar
FROM group_events ge
LEFT JOIN group_rsvp gr ON ge.event_id = gr.event_id
LEFT JOIN users u ON gr.user_id = u.user_id
LEFT JOIN users cu ON ge.creator_id = cu.user_id
WHERE ge.group_id = ?
ORDER BY ge.event_timestamp DESC;

-- name: GetGroupJoinRequests :many
SELECT
    gm.user_id,
    gm.group_id,
    gm.status,
    gm.invited_by,
    gm.created_at,
    u.user_id AS m_user_id,
    u.first_name AS m_first_name,
    u.last_name AS m_last_name
FROM group_members gm
JOIN users u ON gm.user_id = u.user_id
WHERE group_id = ? AND status = 'requested' AND invited_by IS NULL;

-- name: DeleteDbGroup :exec
DELETE FROM groups WHERE group_id = ?;

-- name: UpdateDbGroup :one
UPDATE groups SET group_name = ?, description = ?, image = ?
WHERE group_id = ? AND creator_id = ? RETURNING *;

-- name: GetGroupEvents :many
SELECT e.*,
    (SELECT COUNT(*) FROM group_rsvp r WHERE r.event_id=e.event_id AND r.status='going') as going_count,
    (SELECT COUNT(*) FROM group_rsvp r WHERE r.event_id=e.event_id AND r.status='not going') as not_going_count,
    (SELECT r.status FROM group_rsvp r WHERE r.event_id=e.event_id AND r.user_id=?) as user_rsvp
FROM group_events e
WHERE e.group_id = ?
ORDER BY e.event_timestamp ASC;

-- name: CreateGroupEvent :one
INSERT INTO group_events (event_id, group_id, creator_id, title, description, event_timestamp)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetEventByID :one
SELECT * FROM group_events WHERE event_id = ?;

-- name: UpsertRSVP :exec
INSERT INTO group_rsvp (event_id, user_id, status) VALUES (?, ?, ?)
ON CONFLICT(event_id, user_id) DO UPDATE SET status = excluded.status;

-- name: GetGroupMemberIDs :many
SELECT user_id FROM group_members WHERE group_id = ? AND status = 'joined';

-- name: GetEventGroupID :one
SELECT group_id FROM group_events WHERE event_id = ?;

-- name: CountMembers :one
SELECT count(*) FROM group_members WHERE group_id = ? AND status = 'joined';

