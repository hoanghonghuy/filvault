CREATE TABLE activity_events (
    id CHAR(26) PRIMARY KEY,
    owner_id CHAR(26) NOT NULL REFERENCES users (id),
    type TEXT NOT NULL,
    target_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX activity_events_owner_created
    ON activity_events (owner_id, created_at DESC);
