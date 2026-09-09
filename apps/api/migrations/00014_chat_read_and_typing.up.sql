ALTER TABLE conversation_members
    ADD COLUMN IF NOT EXISTS last_read_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS last_read_message_id CHAR(26) NULL REFERENCES messages (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS conversation_members_read
    ON conversation_members (conversation_id, user_id, last_read_at);
