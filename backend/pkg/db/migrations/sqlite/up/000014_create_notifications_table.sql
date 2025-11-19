-- 000014_create_notifications_table.up.sql
CREATE TABLE IF NOT EXISTS notifications (
    notif_id TEXT PRIMARY KEY,
    receiver_id BLOB NOT NULL,
    type TEXT CHECK(type IN ('follow_request', 'group_invitation', 'group_request', 'group_event', 'message')) NOT NULL,
    is_seen BOOLEAN DEFAULT 0,
    from_id BLOB NOT NULL,
    group_id BLOB,
    event_id BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (receiver_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (from_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES group_events(event_id) ON DELETE CASCADE
);