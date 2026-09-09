DROP INDEX IF EXISTS conversation_members_read;

ALTER TABLE conversation_members
    DROP COLUMN IF EXISTS last_read_message_id,
    DROP COLUMN IF EXISTS last_read_at;
