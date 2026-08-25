CREATE TABLE conversations (
  id CHAR(26) PRIMARY KEY,
  owner_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title TEXT,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  archived_at TIMESTAMPTZ
);

CREATE INDEX idx_conversations_owner_updated
  ON conversations(owner_id, updated_at DESC);

CREATE TABLE messages (
  id CHAR(26) PRIMARY KEY,
  conversation_id CHAR(26) NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  owner_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  body TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_messages_conversation_created
  ON messages(conversation_id, created_at DESC);

CREATE TABLE message_attachments (
  id CHAR(26) PRIMARY KEY,
  message_id CHAR(26) NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  file_id CHAR(26) NOT NULL REFERENCES files(id),
  original_name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE(message_id, file_id)
);
