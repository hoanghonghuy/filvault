ALTER TABLE files
    ADD COLUMN replaces_file_id CHAR(26) NULL REFERENCES files (id);

CREATE TABLE file_versions (
    id CHAR(26) PRIMARY KEY,
    file_id CHAR(26) NOT NULL REFERENCES files (id) ON DELETE CASCADE,
    object_key TEXT NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    mime_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX file_versions_file
    ON file_versions (file_id, created_at DESC);
