-- 000007_create_comments_table.up.sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id BLOB PRIMARY KEY,
    post_id BLOB NOT NULL,
    author_id BLOB NOT NULL,
    parent_comment_id BLOB,
    content TEXT NOT NULL,
    image_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (parent_comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
    FOREIGN KEY (image_id) REFERENCES images(image_id) ON DELETE SET NULL
);