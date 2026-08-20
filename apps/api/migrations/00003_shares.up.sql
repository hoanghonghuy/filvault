CREATE TABLE shares (
    id CHAR(26) PRIMARY KEY,
    owner_id CHAR(26) NOT NULL REFERENCES users (id),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('file', 'folder')),
    resource_id CHAR(26) NOT NULL,
    recipient_id CHAR(26) NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX shares_unique
    ON shares (owner_id, resource_type, resource_id, recipient_id);

CREATE INDEX shares_owner
    ON shares (owner_id);

CREATE INDEX shares_recipient
    ON shares (recipient_id);
