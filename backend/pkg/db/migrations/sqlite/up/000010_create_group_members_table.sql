-- 000010_create_group_members_table.up.sql
CREATE TABLE IF NOT EXISTS group_members (
    user_id BLOB NOT NULL,
    group_id BLOB NOT NULL,
    status TEXT CHECK(status IN ('joined', 'requested')) NOT NULL,
    invited_by BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, group_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (invited_by) REFERENCES users(user_id) ON DELETE SET NULL
);