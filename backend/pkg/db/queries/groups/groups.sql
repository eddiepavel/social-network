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