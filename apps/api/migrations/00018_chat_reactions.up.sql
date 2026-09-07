CREATE TABLE IF NOT EXISTS message_reactions (
    id CHAR(26) PRIMARY KEY,
    message_id CHAR(26) NOT NULL REFERENCES messages (id) ON DELETE CASCADE,
    user_id CHAR(26) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reaction TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (message_id, user_id)
);

CREATE INDEX IF NOT EXISTS message_reactions_message
    ON message_reactions (message_id);

CREATE INDEX IF NOT EXISTS message_reactions_user
    ON message_reactions (user_id);
