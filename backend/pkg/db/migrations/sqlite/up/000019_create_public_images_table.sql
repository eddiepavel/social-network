-- 000019_create_public_images_table.up.sql
CREATE TABLE IF NOT EXISTS public_images (
    image_id TEXT PRIMARY KEY,
    guest_session BLOB NOT NULL,
    image_path TEXT NOT NULL,
    file_name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME
);