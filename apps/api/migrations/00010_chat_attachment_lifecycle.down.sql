DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM message_attachments
        WHERE file_id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot rollback chat attachment lifecycle: purged attachments exist';
    END IF;
END $$;

DROP TABLE IF EXISTS file_purge_jobs;

DROP INDEX IF EXISTS files_completed_message;
DROP INDEX IF EXISTS files_source_ref;
ALTER TABLE message_attachments
    DROP CONSTRAINT IF EXISTS message_attachments_file_id_fkey;

ALTER TABLE message_attachments
    ALTER COLUMN file_id SET NOT NULL;

ALTER TABLE message_attachments
    ADD CONSTRAINT message_attachments_file_id_fkey
        FOREIGN KEY (file_id) REFERENCES files (id);

ALTER TABLE message_attachments
    DROP COLUMN IF EXISTS size_bytes,
    DROP COLUMN IF EXISTS mime_type,
    DROP COLUMN IF EXISTS display_name;

ALTER TABLE files
    DROP COLUMN IF EXISTS source_ref_id,
    DROP COLUMN IF EXISTS completed_message_id,
    DROP COLUMN IF EXISTS source;
