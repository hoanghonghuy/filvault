CREATE TABLE users (
    id CHAR(26) PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    email_verified_at TIMESTAMPTZ NULL,
    verification_code_hash TEXT NULL,
    verification_expires_at TIMESTAMPTZ NULL,
    storage_used BIGINT NOT NULL DEFAULT 0 CHECK (storage_used >= 0),
    storage_quota BIGINT NOT NULL CHECK (storage_quota > 0),
    trash_auto_delete_enabled BOOLEAN NOT NULL,
    trash_retention_days INT NOT NULL CHECK (trash_retention_days >= 1),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE refresh_tokens (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE folders (
    id CHAR(26) PRIMARY KEY,
    owner_id CHAR(26) NOT NULL REFERENCES users (id),
    parent_id CHAR(26) NULL REFERENCES folders (id),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CHECK (parent_id IS DISTINCT FROM id)
);

CREATE UNIQUE INDEX folders_sibling_name
    ON folders (owner_id, parent_id, name)
    WHERE deleted_at IS NULL AND parent_id IS NOT NULL;

CREATE UNIQUE INDEX folders_root_name
    ON folders (owner_id, name)
    WHERE deleted_at IS NULL AND parent_id IS NULL;

CREATE INDEX folders_list_alive
    ON folders (owner_id, parent_id)
    WHERE deleted_at IS NULL;

CREATE INDEX folders_trash
    ON folders (owner_id)
    WHERE deleted_at IS NOT NULL;

CREATE TABLE files (
    id CHAR(26) PRIMARY KEY,
    owner_id CHAR(26) NOT NULL REFERENCES users (id),
    folder_id CHAR(26) NULL REFERENCES folders (id),
    name TEXT NOT NULL,
    original_name TEXT NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    mime_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'READY', 'FAILED', 'PROCESSING', 'UPLOADED')),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    upload_expires_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX files_sibling_name
    ON files (owner_id, folder_id, name)
    WHERE deleted_at IS NULL AND folder_id IS NOT NULL;

CREATE UNIQUE INDEX files_root_name
    ON files (owner_id, name)
    WHERE deleted_at IS NULL AND folder_id IS NULL;

CREATE INDEX files_list_ready
    ON files (owner_id, folder_id)
    WHERE deleted_at IS NULL AND status = 'READY';

CREATE INDEX files_trash
    ON files (owner_id)
    WHERE deleted_at IS NOT NULL;

CREATE INDEX files_pending_expiry
    ON files (status, upload_expires_at)
    WHERE status = 'PENDING';

CREATE INDEX files_photos_timeline
    ON files (owner_id, created_at DESC)
    WHERE deleted_at IS NULL AND status = 'READY';

CREATE TABLE albums (
    id CHAR(26) PRIMARY KEY,
    owner_id CHAR(26) NOT NULL REFERENCES users (id),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX albums_owner_name
    ON albums (owner_id, name);

CREATE TABLE album_items (
    album_id CHAR(26) NOT NULL REFERENCES albums (id) ON DELETE CASCADE,
    file_id CHAR(26) NOT NULL REFERENCES files (id) ON DELETE CASCADE,
    position INT NOT NULL DEFAULT 0,
    added_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (album_id, file_id)
);
