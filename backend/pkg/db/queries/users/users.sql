-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: GetUserByNickname :one
SELECT * FROM users WHERE nickname = ? LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (user_id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
RETURNING *;

-- name: GetUserById :one
SELECT user_id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public, created_at
FROM users
WHERE user_id = ?
LIMIT 1;

-- name: UpdateUser :one
UPDATE users SET first_name = ?, last_name = ?, nickname = ?, about_me = ?, avatar = ? WHERE user_id = ? RETURNING *;

-- name: UpdateUserPrivacy :one
UPDATE users SET is_public = ? WHERE user_id = ? RETURNING *;

-- name: ValidateUserIds :one
SELECT COUNT(*) FROM users WHERE user_id IN (sqlc.slice('user_id'));

-- name: QueryUsers :many
SELECT
    u.user_id,
    u.first_name,
    u.last_name,
    u.nickname
FROM users u
WHERE (
    EXISTS (
        SELECT 1
        FROM followers f
        WHERE 
            (f.follower_id = ?1 AND f.followee_id = u.user_id)
         OR (f.follower_id = u.user_id AND f.followee_id = ?1)
    )
    OR u.is_public = 1
)
AND (
    LOWER(u.first_name || ' ' || u.last_name) LIKE LOWER('%' || ?2 || '%')
    OR LOWER(COALESCE(u.nickname, '')) LIKE LOWER('%' || ?2 || '%')
)