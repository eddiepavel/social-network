-- 000017_add_notification_types.up.sql
-- Add new notification types to CHECK constraint
-- SQLite requires table recreation for CHECK constraint changes

-- Create a new table with updated constraint
CREATE TABLE notifications_new (
    notif_id TEXT PRIMARY KEY,
    receiver_id BLOB NOT NULL,
    type TEXT CHECK(type IN (
        'follow_request',
        'follow_accepted',
        'group_invitation',
        'group_request',
        'group_join_approved',
        'group_join_rejected',
        'group_event',
        'post_comment',
        'comment_reply',
        'post_reaction',
        'comment_reaction',
        'message'
    )) NOT NULL,
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

-- Copy data from old table
INSERT INTO notifications_new SELECT * FROM notifications;

-- Drop old table
DROP TABLE notifications;

-- Rename new table
ALTER TABLE notifications_new RENAME TO notifications;
