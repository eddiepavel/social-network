-- 000004_remove_followers_status.down.sql
ALTER TABLE followers ADD COLUMN status TEXT CHECK(status IN ('pending', 'accepted', 'rejected')) NOT NULL DEFAULT 'pending';
