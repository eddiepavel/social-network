-- name: GetPostWithReactionsAndComments :one
SELECT
    p.*,
    (SELECT COALESCE(json_group_array(json_object(
        'reaction_id', r.reaction_id,
        'user_id', r.user_id,
        'reaction_type', r.reaction_type
                                  )), '[]')
    FROM reactions r WHERE r.target_type = 'post' AND r.target_id = p.post_id) as reactions,
    (SELECT COALESCE(json_group_array(json_object(
        'comment_id', c.comment_id,
        'user_id', c.user_id,
        'content', c.content,
        'reactions', (
            SELECT COALESCE(json_group_array(json_object(
                'reaction_id', cr.reaction_id,
                'user_id', cr.user_id,
                'reaction_type', cr.reaction_type
            )), '[]')
            FROM reactions cr
            WHERE cr.target_type = 'comment' AND cr.target_id = c.comment_id
        )
    )), '[]')
    FROM comments c WHERE c.post_id = p.post_id) as comments
FROM posts p
WHERE p.post_id = ?;

-- name: GetPostVisibility :one
SELECT visibility FROM posts WHERE post_id = ?;

-- name: GetPostBasicInfo :one
SELECT post_id, author_id, visibility FROM posts WHERE post_id = ?;

-- name: CheckPrivatePostUserPermit :one
SELECT * FROM viewing_permissions WHERE user_id = ? AND post_id = ?;

-- name: GetPostsForFeed :many
SELECT p.*,
       i.image_id,
       i.image_path,
       (SELECT COUNT(*)
        FROM reactions
        WHERE target_type = 'post'
          AND target_id = p.post_id) as reaction_count,
       (SELECT COUNT(*)
        FROM comments
        WHERE post_id = p.post_id)   as comment_count
FROM posts p
INNER JOIN users u ON p.author_id = u.user_id
LEFT JOIN images i ON p.image_id = i.image_id
WHERE
   -- Public posts from any user
    (p.visibility = 'public')
   OR
   -- Private posts where current user follows the author
    (p.visibility = 'private' AND EXISTS (SELECT 1
                                          FROM followers f
                                          WHERE f.follower_id = ?
                                            AND f.followee_id = p.author_id))
   OR
   -- Private posts where current user has explicit viewing permission
    (p.visibility = 'private' AND EXISTS (SELECT 1
                                          FROM viewing_permissions vp
                                          WHERE vp.user_id = ?
                                            AND vp.post_id = p.post_id))
ORDER BY p.created_at DESC LIMIT ?
OFFSET ?;

-- name: CreatePost :one
INSERT INTO posts (post_id, author_id, content, visibility, image_id) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: UpdatePost :exec
UPDATE posts SET content = ? AND image_id = ? WHERE post_id = ? and author_id = ?;

-- name: DeletePost :exec
DELETE FROM posts WHERE post_id = ? and author_id = ?;

-- name: EditPostVisibility :exec
UPDATE posts SET visibility = ? WHERE post_id = ? and author_id = ?;

-- name: AddPrivatePostViewingPermission :execrows
INSERT INTO viewing_permissions (user_id, post_id)
SELECT ?, ?
    WHERE EXISTS (
    SELECT 1 FROM posts p
    WHERE p.post_id = ?
      AND p.visibility = 'private'
      AND EXISTS (
          SELECT 1 FROM followers f
          WHERE f.follower_id = ?
            AND f.followee_id = p.author_id
      )
);

-- name: RemovePrivatePostViewingPermission :exec
DELETE FROM viewing_permissions WHERE user_id = ? AND post_id = ?;

-- name: GetPostComments :many
SELECT * FROM comments WHERE post_id = ?;

-- name: GetPostReactions :many
SELECT * FROM reactions WHERE target_type = 'post' AND target_id = ?;

-- name: GetCommentReactions :many
SELECT * FROM reactions WHERE target_type = 'comment' AND target_id = ?;

-- name: CreateComment :execrows
INSERT INTO comments (comment_id, post_id, user_id, content, parent_comment_id, image_id)
SELECT ?, ?, ?, ?, ?, ?
    WHERE EXISTS (
    SELECT 1 FROM posts p
    WHERE p.post_id = ?
      AND (
          p.visibility = 'public'
          OR EXISTS (
              SELECT 1 FROM followers f
              WHERE f.follower_id = ?
                AND f.followee_id = p.author_id
          )
          OR EXISTS (
              SELECT 1 FROM viewing_permissions vp
              WHERE vp.user_id = ?
                AND vp.post_id = p.post_id
          )
      )
);

-- name: DeleteComment :exec
DELETE FROM comments WHERE comment_id = ? AND user_id = ?;

-- name: EditComment :exec
UPDATE comments SET content = ? WHERE comment_id = ? AND user_id = ?;

-- name: CreateReaction :execrows
INSERT INTO reactions (reaction_id, target_type, target_id, user_id, reaction_type)
SELECT ?, ?, ?, ?, ?
    WHERE EXISTS (
    SELECT 1 FROM posts p
    WHERE p.post_id = (
        CASE
            WHEN ? = 'post' THEN ?
            WHEN ? = 'comment' THEN (SELECT post_id FROM comments WHERE comment_id = ?)
        END
    )
    AND (
        p.visibility = 'public'
        OR EXISTS (
            SELECT 1 FROM followers f
            WHERE f.follower_id = ?
              AND f.followee_id = p.author_id
        )
        OR EXISTS (
            SELECT 1 FROM viewing_permissions vp
            WHERE vp.user_id = ?
              AND vp.post_id = p.post_id
        )
    )
);

-- name: DeleteReaction :exec
DELETE FROM reactions WHERE reaction_id = ? AND user_id = ?;