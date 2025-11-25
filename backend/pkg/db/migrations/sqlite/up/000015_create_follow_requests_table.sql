-- 000015_create_follow_requests_table.sql
CREATE TABLE IF NOT EXISTS follow_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    follower_id BLOB NOT NULL,
    followee_id BLOB NOT NULL,
    status VARCHAR(20) CHECK (status IN ('pending', 'accepted', 'declined')) DEFAULT 'pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at DATETIME NULL,

    FOREIGN KEY (follower_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (followee_id) REFERENCES users(user_id) ON DELETE CASCADE
);