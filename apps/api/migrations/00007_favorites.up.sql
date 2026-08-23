CREATE TABLE favorites (
    id CHAR(26) PRIMARY KEY,
    owner_id CHAR(26) NOT NULL REFERENCES users(id),
    file_id CHAR(26) NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (owner_id, file_id)
);

CREATE INDEX favorites_owner_created
    ON favorites (owner_id, created_at DESC);
