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