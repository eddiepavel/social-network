-- 000008_create_reactions_table.up.sql
CREATE TABLE IF NOT EXISTS reactions (
    reaction_id BLOB PRIMARY KEY,
    author_id BLOB NOT NULL,
    target_type TEXT CHECK(target_type IN ('post', 'comment')) NOT NULL,
    target_id BLOB NOT NULL,
    reacted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE CASCADE
);