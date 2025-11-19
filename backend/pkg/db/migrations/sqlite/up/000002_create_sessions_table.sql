-- 000002_create_sessions_table.up.sql
CREATE TABLE IF NOT EXISTS sessions (
    session_id BLOB PRIMARY KEY,
    user_id BLOB NOT NULL,
    active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);