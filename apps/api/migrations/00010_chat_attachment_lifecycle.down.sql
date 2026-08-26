DROP TABLE file_purge_jobs;

DROP INDEX files_completed_message;
DROP INDEX files_source_ref;
ALTER TABLE message_attachments
    DROP CONSTRAINT IF EXISTS message_attachments_file_id_fkey;

ALTER TABLE message_attachments
    ALTER COLUMN file_id SET NOT NULL;

ALTER TABLE message_attachments
    ADD CONSTRAINT message_attachments_file_id_fkey
        FOREIGN KEY (file_id) REFERENCES files (id);

ALTER TABLE message_attachments
    DROP COLUMN size_bytes,
    DROP COLUMN mime_type,
    DROP COLUMN display_name;

ALTER TABLE files
    DROP COLUMN source_ref_id,
    DROP COLUMN completed_message_id,
    DROP COLUMN source;
