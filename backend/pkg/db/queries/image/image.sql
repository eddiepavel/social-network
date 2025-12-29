-- name: CreateImage :exec
INSERT INTO images (image_id, poster_id, image_path, created_at, expires_at)
VALUES(?, ?, ?, ?, ?);

-- name: ImageState :exec
UPDATE images SET expires_at = ? WHERE image_id = ?; 

-- name: GetNotSetImages :many
SELECT * FROM images WHERE expires_at IS NOT NULL; 

-- name: DeleteImages :exec
DELETE FROM images WHERE image_id IN (sqlc.slice('image_id'));

-- name: GetImageById :one
SELECT * FROM images WHERE image_id = ? LIMIT 1;