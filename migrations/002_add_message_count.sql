ALTER TABLE users ADD COLUMN IF NOT EXISTS message_count INTEGER DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_users_message_count ON users(message_count);
