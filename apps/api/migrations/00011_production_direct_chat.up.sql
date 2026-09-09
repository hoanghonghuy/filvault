ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS conversation_type TEXT NOT NULL DEFAULT 'legacy'
        CHECK (conversation_type IN ('legacy', 'direct')),
    ADD COLUMN IF NOT EXISTS direct_key TEXT NULL,
    ADD COLUMN IF NOT EXISTS created_by_user_id CHAR(26) NULL REFERENCES users (id),
    ADD COLUMN IF NOT EXISTS last_message_at TIMESTAMPTZ NULL;

UPDATE conversations
SET created_by_user_id = owner_id
WHERE created_by_user_id IS NULL;

UPDATE conversations
SET last_message_at = updated_at
WHERE last_message_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS conversations_direct_key
    ON conversations (direct_key)
    WHERE conversation_type = 'direct' AND direct_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS conversation_members (
    conversation_id CHAR(26) NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,
    user_id CHAR(26) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ NULL,
    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX IF NOT EXISTS conversation_members_inbox
    ON conversation_members (user_id, archived_at, conversation_id);

INSERT INTO conversation_members (conversation_id, user_id, joined_at)
SELECT id, owner_id, created_at
FROM conversations
ON CONFLICT (conversation_id, user_id) DO NOTHING;

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS sender_id CHAR(26) NULL REFERENCES users (id),
    ADD COLUMN IF NOT EXISTS client_message_id TEXT NULL,
    ADD COLUMN IF NOT EXISTS edited_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS removed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS removed_by_user_id CHAR(26) NULL REFERENCES users (id);

UPDATE messages
SET sender_id = owner_id
WHERE sender_id IS NULL;

ALTER TABLE messages
    ALTER COLUMN sender_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS messages_sender_client_id
    ON messages (conversation_id, sender_id, client_message_id)
    WHERE client_message_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS messages_conversation_created_id
    ON messages (conversation_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS chat_shared_albums (
    id CHAR(26) PRIMARY KEY,
    conversation_id CHAR(26) NOT NULL UNIQUE REFERENCES conversations (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS chat_shared_album_items (
    album_id CHAR(26) NOT NULL REFERENCES chat_shared_albums (id) ON DELETE CASCADE,
    attachment_id CHAR(26) NOT NULL UNIQUE REFERENCES message_attachments (id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL,
    hidden_at TIMESTAMPTZ NULL,
    PRIMARY KEY (album_id, attachment_id)
);

INSERT INTO chat_shared_albums (id, conversation_id, name, created_at, updated_at)
SELECT c.id, c.id, 'Shared media', c.created_at, c.updated_at
FROM conversations c
WHERE EXISTS (
    SELECT 1
    FROM messages m
    JOIN message_attachments ma ON ma.message_id = m.id
    WHERE m.conversation_id = c.id
)
ON CONFLICT (conversation_id) DO NOTHING;

INSERT INTO chat_shared_album_items (album_id, attachment_id, added_at)
SELECT c.id, ma.id, ma.created_at
FROM message_attachments ma
JOIN messages m ON m.id = ma.message_id
JOIN conversations c ON c.id = m.conversation_id
ON CONFLICT (attachment_id) DO NOTHING;

CREATE INDEX IF NOT EXISTS chat_shared_album_items_visible
    ON chat_shared_album_items (album_id, added_at DESC)
    WHERE hidden_at IS NULL;

CREATE TABLE IF NOT EXISTS chat_events (
    id CHAR(26) PRIMARY KEY,
    conversation_id CHAR(26) NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    aggregate_id CHAR(26) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    sequence BIGSERIAL NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS chat_events_conversation_sequence
    ON chat_events (conversation_id, sequence);

CREATE INDEX IF NOT EXISTS chat_events_sequence
    ON chat_events (sequence);

ALTER TABLE file_purge_jobs
    ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS claimed_by TEXT NULL,
    ADD COLUMN IF NOT EXISTS last_error TEXT NULL,
    ADD COLUMN IF NOT EXISTS failed_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS file_purge_jobs_ready
    ON file_purge_jobs (next_attempt_at, created_at)
    WHERE completed_at IS NULL AND failed_at IS NULL;

