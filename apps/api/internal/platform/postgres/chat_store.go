package postgres

import (
	"context"
	"errors"

	"filvault/internal/chat"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateConversation(ctx context.Context, c chat.Conversation) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO conversations (id, owner_id, title, created_at, updated_at, archived_at)
		VALUES ($1, $2, $3, $4, $5, NULL)
	`, c.ID, c.OwnerID, c.Title, c.CreatedAt, c.UpdatedAt)
	return err
}

func (s *Store) ListConversations(ctx context.Context, ownerID string) ([]chat.Conversation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, owner_id, COALESCE(title, ''), created_at, updated_at
		FROM conversations
		WHERE owner_id = $1 AND archived_at IS NULL
		ORDER BY updated_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []chat.Conversation
	for rows.Next() {
		var c chat.Conversation
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetConversation(ctx context.Context, ownerID, id string) (*chat.Conversation, error) {
	var c chat.Conversation
	err := s.pool.QueryRow(ctx, `
		SELECT id, owner_id, COALESCE(title, ''), created_at, updated_at
		FROM conversations
		WHERE owner_id = $1 AND id = $2 AND archived_at IS NULL
	`, ownerID, id).Scan(&c.ID, &c.OwnerID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) CreateMessage(ctx context.Context, m chat.Message) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO messages (id, conversation_id, owner_id, body, created_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, NULL)
	`, m.ID, m.ConversationID, m.OwnerID, m.Body, m.CreatedAt)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE conversations SET updated_at = $3
		WHERE id = $1 AND owner_id = $2
	`, m.ConversationID, m.OwnerID, m.CreatedAt)
	return err
}

func (s *Store) CreateAttachment(ctx context.Context, a chat.Attachment) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO message_attachments (id, message_id, file_id, original_name, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, a.ID, a.MessageID, a.FileID, a.OriginalName, a.CreatedAt)
	return err
}

func (s *Store) ListMessages(ctx context.Context, ownerID, conversationID string, limit int) ([]chat.Message, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, conversation_id, owner_id, body, created_at
		FROM messages
		WHERE owner_id = $1 AND conversation_id = $2 AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $3
	`, ownerID, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []chat.Message
	for rows.Next() {
		var m chat.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.OwnerID, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		attachments, err := s.listMessageAttachments(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Attachments = attachments
	}
	return out, nil
}

func (s *Store) listMessageAttachments(ctx context.Context, messageID string) ([]chat.Attachment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.message_id, a.file_id, a.original_name, f.name, f.mime_type, f.size_bytes, a.created_at
		FROM message_attachments a
		JOIN files f ON f.id = a.file_id
		WHERE a.message_id = $1
		ORDER BY a.created_at ASC
	`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []chat.Attachment
	for rows.Next() {
		var a chat.Attachment
		if err := rows.Scan(&a.ID, &a.MessageID, &a.FileID, &a.OriginalName, &a.Name, &a.MimeType, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type chatRepo struct {
	store *Store
}

func NewChatRepository(store *Store) chat.Repository {
	return chatRepo{store: store}
}

func (r chatRepo) CreateConversation(ctx context.Context, c chat.Conversation) error {
	return r.store.CreateConversation(ctx, c)
}

func (r chatRepo) ListConversations(ctx context.Context, ownerID string) ([]chat.Conversation, error) {
	return r.store.ListConversations(ctx, ownerID)
}

func (r chatRepo) GetConversation(ctx context.Context, ownerID, id string) (*chat.Conversation, error) {
	return r.store.GetConversation(ctx, ownerID, id)
}

func (r chatRepo) CreateMessage(ctx context.Context, m chat.Message) error {
	return r.store.CreateMessage(ctx, m)
}

func (r chatRepo) ListMessages(ctx context.Context, ownerID, conversationID string, limit int) ([]chat.Message, error) {
	return r.store.ListMessages(ctx, ownerID, conversationID, limit)
}

func (r chatRepo) CreateAttachment(ctx context.Context, a chat.Attachment) error {
	return r.store.CreateAttachment(ctx, a)
}

var _ chat.Repository = chatRepo{}
