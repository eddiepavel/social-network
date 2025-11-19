-- 000011_create_group_events_table.up.sql
CREATE TABLE IF NOT EXISTS group_events (
    event_id TEXT PRIMARY KEY,
    creator_id BLOB NOT NULL,
    group_id BLOB NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    event_timestamp DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE
);