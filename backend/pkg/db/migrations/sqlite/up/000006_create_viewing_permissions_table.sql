-- 000006_create_viewing_permissions_table.up.sql
CREATE TABLE IF NOT EXISTS viewing_permissions (
    post_id TEXT NOT NULL,
    user_id BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);