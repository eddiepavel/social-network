-- name: GetFollowRequests :many
SELECT * FROM follow_requests WHERE followee_id = ? AND status = 'pending';

-- name: GetFollowRequestByID :one
SELECT * FROM follow_requests WHERE id = ? AND followee_id = ? AND status = 'pending';

-- name: GetFollowRequest :one
SELECT * FROM follow_requests WHERE follower_id = ? AND followee_id = ?;

-- name: DeleteFollowRequest :exec
DELETE FROM follow_requests WHERE id = ?;

-- name: AcceptFollowRequest :exec
UPDATE follow_requests SET status = 'accepted' WHERE id = ?;

-- name: CreateFollowRequest :exec
INSERT INTO follow_requests (follower_id, followee_id) VALUES (?, ?);