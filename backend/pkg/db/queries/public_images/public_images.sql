-- name: SavePublicImage :exec
INSERT INTO public_images (image_id, guest_session, image_path, file_name, created_at, expires_at)
VALUES(?, ?, ?, ?, CAST(? AS TEXT), CAST(? AS TEXT));

-- name: DeletePublicImage :exec
DELETE FROM public_images WHERE image_id = ?;

-- name: GetPublicImage :one
SELECT * FROM public_images WHERE image_id = ? LIMIT 1;

-- name: GetPublicImages :many
SELECT * FROM public_images;

-- name: DeletePublicImages :exec
DELETE FROM public_images WHERE image_id IN (sqlc.slice('image_id'));