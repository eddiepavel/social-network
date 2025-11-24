-- name: CheckIfUserFollows :one
SELECT * FROM followers WHERE follower_id = ? AND followee_id = ?;