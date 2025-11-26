-- 000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    user_id BLOB PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash BLOB NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    dob TEXT NOT NULL,
    avatar TEXT,
    nickname TEXT,
    about_me TEXT,
    is_public BOOLEAN DEFAULT 0 NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);