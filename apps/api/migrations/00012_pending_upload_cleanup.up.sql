CREATE TABLE IF NOT EXISTS file_upload_cleanup_jobs (
    id CHAR(26) PRIMARY KEY,
    owner_id CHAR(26) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    file_id CHAR(26) NOT NULL UNIQUE,
    object_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ NULL,
    attempts INT NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    claimed_at TIMESTAMPTZ NULL,
    claimed_by TEXT NULL,
    last_error TEXT NULL,
    failed_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS file_upload_cleanup_jobs_ready
    ON file_upload_cleanup_jobs (next_attempt_at, created_at)
    WHERE completed_at IS NULL AND failed_at IS NULL;

INSERT INTO file_upload_cleanup_jobs (id, owner_id, file_id, object_key, created_at)
SELECT f.id, f.owner_id, f.id, f.object_key, COALESCE(f.upload_expires_at, f.created_at)
FROM files f
WHERE f.status = 'PENDING'
    AND f.upload_expires_at IS NOT NULL
    AND f.upload_expires_at <= now()
ON CONFLICT (file_id) DO NOTHING;
