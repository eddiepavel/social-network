-- 000009_create_groups_table.up.sql
CREATE TABLE IF NOT EXISTS groups (
    group_id BLOB PRIMARY KEY,
    group_name TEXT NOT NULL,
    description TEXT NOT NULL,
    image TEXT,
    creator_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (image) REFERENCES images(image_id) ON DELETE SET NULL
);