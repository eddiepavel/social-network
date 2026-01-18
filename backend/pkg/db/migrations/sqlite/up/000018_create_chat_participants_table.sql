-- 000018_create_chat_participants.sql
CREATE TABLE IF NOT EXISTS chat_participants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id BLOB NOT NULL,
    user_id BLOB NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_read_at TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES chat_rooms(room_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    UNIQUE (room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_participants_user
    ON chat_participants(user_id, room_id);