DROP INDEX IF EXISTS file_purge_jobs_ready;

ALTER TABLE file_purge_jobs
    DROP COLUMN IF EXISTS failed_at,
    DROP COLUMN IF EXISTS last_error,
    DROP COLUMN IF EXISTS claimed_by,
    DROP COLUMN IF EXISTS claimed_at,
    DROP COLUMN IF EXISTS next_attempt_at,
    DROP COLUMN IF EXISTS attempts;

DROP TABLE IF EXISTS chat_events;
DROP TABLE IF EXISTS chat_shared_album_items;
DROP TABLE IF EXISTS chat_shared_albums;

DROP INDEX IF EXISTS messages_conversation_created_id;
DROP INDEX IF EXISTS messages_sender_client_id;

ALTER TABLE messages
    DROP COLUMN IF EXISTS removed_by_user_id,
    DROP COLUMN IF EXISTS removed_at,
    DROP COLUMN IF EXISTS edited_at,
    DROP COLUMN IF EXISTS client_message_id,
    DROP COLUMN IF EXISTS sender_id;

DROP INDEX IF EXISTS conversation_members_inbox;
DROP TABLE IF EXISTS conversation_members;

DROP INDEX IF EXISTS conversations_direct_key;

ALTER TABLE conversations
    DROP COLUMN IF EXISTS last_message_at,
    DROP COLUMN IF EXISTS created_by_user_id,
    DROP COLUMN IF EXISTS direct_key,
    DROP COLUMN IF EXISTS conversation_type;
