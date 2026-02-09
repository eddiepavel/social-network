-- 000017_add_notification_types.down.sql
-- Revert to original 5 notification types

-- Create a new table with original constraint
CREATE TABLE notifications_new (
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

-- Copy only valid notification types
INSERT INTO notifications_new
SELECT * FROM notifications
WHERE type IN ('follow_request', 'group_invitation', 'group_request', 'group_event', 'message');

-- Drop old table
DROP TABLE notifications;

-- Rename new table
ALTER TABLE notifications_new RENAME TO notifications;
