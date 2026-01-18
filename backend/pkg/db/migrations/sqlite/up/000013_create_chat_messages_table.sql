-- 000013_create_chat_table.up.sql
CREATE TABLE IF NOT EXISTS chat_messages (
    message_id BLOB PRIMARY KEY,
    content TEXT NOT NULL,
    sender_id BLOB NOT NULL,
    target_id BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES chat_rooms(room_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_room_time
    ON chat_messages(target_id, created_at DESC);