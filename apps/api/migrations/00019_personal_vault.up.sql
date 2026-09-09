CREATE TABLE IF NOT EXISTS user_vaults (
    user_id CHAR(26) PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    pin_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE files ADD COLUMN IF NOT EXISTS is_vault BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS files_vault_list
    ON files (owner_id, created_at DESC)
    WHERE deleted_at IS NULL AND is_vault = TRUE AND status = 'READY';
