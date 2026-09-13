ALTER TABLE users
    ADD COLUMN password_reset_token_hash TEXT NULL,
    ADD COLUMN password_reset_expires_at TIMESTAMPTZ NULL;

CREATE UNIQUE INDEX users_password_reset_token_hash
    ON users (password_reset_token_hash)
    WHERE password_reset_token_hash IS NOT NULL;
