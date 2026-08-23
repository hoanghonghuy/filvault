CREATE TABLE share_links (
    id CHAR(26) PRIMARY KEY,
    owner_id CHAR(26) NOT NULL REFERENCES users (id),
    file_id CHAR(26) NOT NULL REFERENCES files (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL,
    recipient_user_id CHAR(26) NULL REFERENCES users (id)
);

CREATE INDEX share_links_owner_created
    ON share_links (owner_id, created_at DESC);
