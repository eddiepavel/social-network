-- name: CheckIfUserFollows :one
SELECT * FROM followers WHERE follower_id = ? AND followee_id = ?;

-- name: InsertFollower :exec
INSERT INTO followers (follower_id, followee_id) VALUES (?, ?);

-- name: DeleteFollower :exec
DELETE FROM followers WHERE follower_id = ? AND followee_id = ?;

-- name: GetFollowers :many
SELECT * FROM followers WHERE followee_id = ?;

-- name: GetFollowees :many
SELECT * FROM followers WHERE follower_id = ?;

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

-- name: CheckPendingFollowRequest :one
SELECT * FROM follow_requests
WHERE follower_id = ? AND followee_id = ? AND status = 'pending';