-- 000012_create_group_rsvp_table.up.sql
CREATE TABLE IF NOT EXISTS group_rsvp (
    event_id BLOB NOT NULL,
    user_id BLOB NOT NULL,
    status TEXT CHECK(status IN ('going', 'not going')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, user_id),
    FOREIGN KEY (event_id) REFERENCES group_events(event_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);