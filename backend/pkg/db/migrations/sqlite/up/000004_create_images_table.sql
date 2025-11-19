-- 000004_create_images_table.up.sql
CREATE TABLE IF NOT EXISTS images (
    image_id TEXT PRIMARY KEY,
    poster_id TEXT NOT NULL,
    image_path TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (poster_id) REFERENCES users(user_id) ON DELETE CASCADE
);