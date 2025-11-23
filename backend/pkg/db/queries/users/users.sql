-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: GetUserByNickname :one
SELECT * FROM users WHERE nickname = ? LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (user_id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
RETURNING *;

-- name: GerUserById :one 
SELECT user_id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public, created_at
FROM users
WHERE user_id = ?
LIMIT 1; 