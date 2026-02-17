-- 000017_create_chat_rooms_table.sql
CREATE TABLE IF NOT EXISTS chat_rooms (
    room_id BLOB PRIMARY KEY,
    name TEXT,
    group_id BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE
);
