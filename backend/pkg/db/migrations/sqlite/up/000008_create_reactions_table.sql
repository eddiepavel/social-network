-- 000008_create_reactions_table.up.sql
CREATE TABLE IF NOT EXISTS reactions (
    reaction_id TEXT PRIMARY KEY,
    user_id BLOB NOT NULL,
    target_type TEXT CHECK(target_type IN ('post', 'comment')) NOT NULL,
    target_id TEXT NOT NULL,
    reaction_type TEXT NOT NULL,
    reacted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);