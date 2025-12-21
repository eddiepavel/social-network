-- 000005_create_posts_table.up.sql
CREATE TABLE IF NOT EXISTS posts (
    post_id BLOB PRIMARY KEY,
    author_id BLOB NOT NULL,
    content TEXT NOT NULL,
    image_id TEXT,
    visibility TEXT CHECK(visibility IN ('public', 'semi-private', 'private')) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (image_id) REFERENCES images(image_id) ON DELETE SET NULL
);