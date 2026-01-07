-- name: GetPostByID :one
SELECT * FROM posts WHERE post_id = ?;

-- name: GetPostVisibility :one
SELECT visibility FROM posts WHERE post_id = ?;

-- name: GetPostBasicInfo :one
SELECT post_id, author_id, visibility FROM posts WHERE post_id = ?;

-- name: CheckPrivatePostUserPermit :one
SELECT * FROM viewing_permissions WHERE user_id = ? AND post_id = ?;

-- name: GetFeedPostsCount :one
SELECT COUNT(*)
FROM posts p
         INNER JOIN users u ON p.author_id = u.user_id
WHERE
    (p.author_id = ?)
   OR
   -- Public posts from any user
    (p.visibility = 'public')
   OR
   -- Private posts where current user follows the author
    (p.visibility = 'semi-private' AND EXISTS (SELECT 1
                                               FROM followers f
                                               WHERE f.follower_id = ?
                                                 AND f.followee_id = p.author_id))
   OR
   -- Private posts where current user has explicit viewing permission
    (p.visibility = 'private' AND EXISTS (SELECT 1
                                          FROM viewing_permissions vp
                                          WHERE vp.user_id = ?
                                            AND vp.post_id = p.post_id));

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
        WHERE post_id = p.post_id)   as comment_count,
       EXISTS(SELECT 1
              FROM reactions r
              WHERE r.target_type = 'post'
                AND r.target_id = p.post_id
                AND r.author_id = ?) as user_reacted
FROM posts p
INNER JOIN users u ON p.author_id = u.user_id
LEFT JOIN images i ON p.image_id = i.image_id
WHERE
    (p.author_id = ?)
    OR
   -- Public posts from any user
    (p.visibility = 'public')
   OR
   -- Private posts where current user follows the author
    (p.visibility = 'semi-private' AND EXISTS (SELECT 1
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
UPDATE posts SET content = ?, image_id = ? WHERE post_id = ? and author_id = ?;

-- name: DeletePost :exec
DELETE FROM posts WHERE post_id = ? and author_id = ?;

-- name: EditPostVisibility :exec
UPDATE posts SET visibility = ? WHERE post_id = ? and author_id = ?;

-- name: PostVisibilitySemiPrivateBatch :exec
UPDATE posts SET visibility = 'semi-private' WHERE author_id = ? and visibility = 'public';

-- name: PostVisibilityPublicBatch :exec
UPDATE posts SET visibility = 'public' WHERE author_id = ? and visibility = 'semi-private';

-- name: RemoveViewingPermissionPostIDBatch :exec
DELETE FROM viewing_permissions WHERE post_id = ?;

-- name: RemoveViewingPermissionUserIDBatch :exec
DELETE FROM viewing_permissions
WHERE user_id = ?
  AND post_id IN (
    SELECT post_id FROM posts WHERE author_id = ?
);

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
SELECT
    c.*,
    COUNT(r.reaction_id) as reaction_count,
    EXISTS(
        SELECT 1 FROM reactions r
        WHERE target_type = 'comment'
          AND r.target_id = c.comment_id
          AND r.author_id = ?
    ) as user_reacted
FROM comments c
         LEFT JOIN reactions r ON r.target_type = 'comment' AND r.target_id = c.comment_id
WHERE c.post_id = ?
GROUP BY c.comment_id;

-- name: GetPostReactions :one
SELECT COUNT(*) FROM reactions WHERE target_type = 'post' AND target_id = ?;

-- name: CreateComment :execrows
INSERT INTO comments (comment_id, post_id, author_id, content, parent_comment_id, image_id)
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
DELETE FROM comments WHERE comment_id = ? AND author_id = ?;

-- name: EditComment :exec
UPDATE comments SET content = ? WHERE comment_id = ? AND author_id = ?;

-- name: CreateReaction :execrows
INSERT INTO reactions (reaction_id, target_type, target_id, author_id)
SELECT ?, ?, ?, ?
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

-- name: HasUserReacted :one
SELECT COUNT(*) FROM reactions WHERE author_id = ? AND target_type = ? AND target_id = ? LIMIT 1;

-- name: FindUserReaction :one
SELECT reaction_id FROM reactions WHERE author_id = ? AND target_type = ? AND target_id = ?;

-- name: DeleteReaction :exec
DELETE FROM reactions WHERE author_id = ? AND target_type = ? AND target_id = ?;

-- name: CheckCommentExists :one
SELECT EXISTS(SELECT 1 FROM comments WHERE comment_id = ?);