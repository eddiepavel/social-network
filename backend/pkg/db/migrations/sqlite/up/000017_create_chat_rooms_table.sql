-- 000017_create_chat_rooms_table.sql
CREATE TABLE IF NOT EXISTS chat_rooms (
    room_id BLOB PRIMARY KEY,
    name TEXT,
    is_group INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
