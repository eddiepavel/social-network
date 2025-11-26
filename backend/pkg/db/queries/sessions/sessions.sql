-- name: GetSessionByUserID :one
SELECT * FROM sessions WHERE user_id = ?;

-- name: CreateSession :one
INSERT INTO sessions (session_id, user_id, active, expires_at) VALUES (?, ?, 1, ?) RETURNING *;