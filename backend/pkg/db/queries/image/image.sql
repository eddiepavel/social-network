-- name: CreateImage :exec
INSERT INTO images (image_id, poster_id, image_path, file_name, created_at, expires_at)
VALUES(?, ?, ?, ?, CAST(? AS TEXT), CAST(? AS TEXT));

-- name: AssignImage :exec
UPDATE images SET expires_at = NULL WHERE image_id = ?;

-- name: SetImageExpiry :exec
UPDATE images SET expires_at = CAST(? AS TEXT) WHERE image_id = ?; 

-- name: GetNotSetImages :many
SELECT * FROM images WHERE expires_at IS NOT NULL AND expires_at != ''; 

-- name: DeleteImages :exec
DELETE FROM images WHERE image_id IN (sqlc.slice('image_id'));

-- name: GetImageById :one
SELECT * FROM images WHERE image_id = ? LIMIT 1;