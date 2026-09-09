ALTER TABLE files
    ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'vault'
        CHECK (source IN ('vault', 'chat')),
    ADD COLUMN IF NOT EXISTS source_ref_id CHAR(26) NULL,
    ADD COLUMN IF NOT EXISTS completed_message_id CHAR(26) NULL;

UPDATE files f
SET source = 'chat',
    source_ref_id = m.conversation_id
FROM message_attachments a
JOIN messages m ON m.id = a.message_id
WHERE a.file_id = f.id;

ALTER TABLE message_attachments
    ADD COLUMN IF NOT EXISTS display_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mime_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    ADD COLUMN IF NOT EXISTS size_bytes BIGINT NOT NULL DEFAULT 0 CHECK (size_bytes >= 0);

UPDATE message_attachments a
SET display_name = f.name,
    mime_type = f.mime_type,
    size_bytes = f.size_bytes
FROM files f
WHERE a.file_id = f.id;

ALTER TABLE message_attachments
    ALTER COLUMN file_id DROP NOT NULL;

ALTER TABLE message_attachments
    DROP CONSTRAINT IF EXISTS message_attachments_file_id_fkey;

ALTER TABLE message_attachments
    ADD CONSTRAINT message_attachments_file_id_fkey
        FOREIGN KEY (file_id) REFERENCES files (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS files_source_ref
    ON files (owner_id, source, source_ref_id)
    WHERE source_ref_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS files_completed_message
    ON files (completed_message_id)
    WHERE completed_message_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS file_purge_jobs (
    id CHAR(26) PRIMARY KEY,
    owner_id CHAR(26) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    file_id CHAR(26) NOT NULL,
    object_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS file_purge_jobs_pending_file
    ON file_purge_jobs (owner_id, file_id)
    WHERE completed_at IS NULL;

CREATE INDEX IF NOT EXISTS file_purge_jobs_pending
    ON file_purge_jobs (created_at)
    WHERE completed_at IS NULL;
