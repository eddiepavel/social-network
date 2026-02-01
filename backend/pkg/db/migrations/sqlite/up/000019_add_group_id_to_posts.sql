-- 000019_add_group_id_to_posts.sql
ALTER TABLE posts ADD COLUMN group_id BLOB REFERENCES groups(group_id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_posts_group_id ON posts(group_id);
