-- 000013_create_chat_table.up.sql
CREATE TABLE IF NOT EXISTS chat (
    message_id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    sender_id BLOB NOT NULL,
    receiver_id BLOB,
    group_id BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (receiver_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    CHECK ((receiver_id IS NOT NULL AND group_id IS NULL) OR (receiver_id IS NULL AND group_id IS NOT NULL))
);