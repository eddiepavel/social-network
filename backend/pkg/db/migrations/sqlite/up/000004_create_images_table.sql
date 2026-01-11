-- 000004_create_images_table.up.sql
CREATE TABLE IF NOT EXISTS images (
    image_id TEXT PRIMARY KEY,
    poster_id BLOB NOT NULL,
    image_path TEXT NOT NULL,
    file_name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    FOREIGN KEY (poster_id) REFERENCES users(user_id) ON DELETE CASCADE
);