package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/chat"
	"filvault/internal/file"
	"filvault/internal/platform/objectstore"

	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
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

func (s *Store) CreateMessageWithAttachments(ctx context.Context, m chat.Message, fileIDs []string) (chat.Message, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return chat.Message{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO messages (id, conversation_id, owner_id, body, created_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, NULL)
	`, m.ID, m.ConversationID, m.OwnerID, m.Body, m.CreatedAt); err != nil {
		return chat.Message{}, err
	}
	seen := make(map[string]struct{}, len(fileIDs))
	for _, fileID := range fileIDs {
		if _, ok := seen[fileID]; ok {
			continue
		}
		seen[fileID] = struct{}{}
		var f file.File
		err := tx.QueryRow(ctx, `
			SELECT id, owner_id, folder_id, name, original_name, object_key,
				mime_type, size_bytes, status, created_at, updated_at, deleted_at,
				upload_expires_at, replaces_file_id, source, source_ref_id
			FROM files
			WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL AND status = 'READY'
		`, fileID, m.OwnerID).Scan(
			&f.ID, &f.OwnerID, &f.FolderID, &f.Name, &f.OriginalName, &f.ObjectKey,
			&f.MimeType, &f.SizeBytes, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.DeletedAt,
			&f.UploadExpiresAt, &f.ReplacesFileID, &f.Source, &f.SourceRefID,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return chat.Message{}, apperr.NotFound
		}
		if err != nil {
			return chat.Message{}, err
		}
		a := chat.Attachment{
			ID:           ulid.Make().String(),
			MessageID:    m.ID,
			FileID:       &f.ID,
			OriginalName: f.OriginalName,
			Name:         f.Name,
			MimeType:     f.MimeType,
			SizeBytes:    f.SizeBytes,
			CreatedAt:    m.CreatedAt,
			Availability: chat.AttachmentAvailable,
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO message_attachments
				(id, message_id, file_id, original_name, display_name, mime_type, size_bytes, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, a.ID, a.MessageID, a.FileID, a.OriginalName, a.Name, a.MimeType, a.SizeBytes, a.CreatedAt); err != nil {
			return chat.Message{}, err
		}
		m.Attachments = append(m.Attachments, a)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE conversations SET updated_at = $3
		WHERE id = $1 AND owner_id = $2
	`, m.ConversationID, m.OwnerID, m.CreatedAt); err != nil {
		return chat.Message{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return chat.Message{}, err
	}
	return m, nil
}

func (s *Store) CompleteAttachment(ctx context.Context, ownerID, conversationID string, pending file.File, stat objectstore.ObjectStat, body string, now time.Time) (chat.Message, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return chat.Message{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	var completedMessageID *string
	var sourceRefID *string
	err = tx.QueryRow(ctx, `
		SELECT status, completed_message_id, source_ref_id
		FROM files
		WHERE id = $1 AND owner_id = $2
		FOR UPDATE
	`, pending.ID, ownerID).Scan(&status, &completedMessageID, &sourceRefID)
	if errors.Is(err, pgx.ErrNoRows) {
		return chat.Message{}, apperr.NotFound
	}
	if err != nil {
		return chat.Message{}, err
	}
	if status == file.StatusReady && completedMessageID != nil {
		var m chat.Message
		err := tx.QueryRow(ctx, `
			SELECT id, conversation_id, owner_id, body, created_at
			FROM messages
			WHERE id = $1 AND owner_id = $2
		`, *completedMessageID, ownerID).Scan(&m.ID, &m.ConversationID, &m.OwnerID, &m.Body, &m.CreatedAt)
		if err != nil {
			return chat.Message{}, err
		}
		if m.ConversationID != conversationID {
			return chat.Message{}, apperr.NotFound
		}
		m.Attachments, err = s.listMessageAttachmentsFrom(ctx, tx, m.ID)
		if err != nil {
			return chat.Message{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return chat.Message{}, err
		}
		return m, nil
	}
	if status == file.StatusReady {
		return chat.Message{}, apperr.InvalidState
	}
	if status != file.StatusPending || sourceRefID == nil || *sourceRefID != conversationID {
		return chat.Message{}, apperr.InvalidState
	}

	var conversationOwner string
	if err := tx.QueryRow(ctx, `
		SELECT owner_id FROM conversations WHERE id = $1 AND owner_id = $2 AND archived_at IS NULL FOR UPDATE
	`, conversationID, ownerID).Scan(&conversationOwner); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return chat.Message{}, apperr.NotFound
		}
		return chat.Message{}, err
	}
	var storageUsed int64
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET storage_used = storage_used + $2, updated_at = now()
		WHERE id = $1 AND storage_used + $2 <= storage_quota
		RETURNING storage_used
	`, ownerID, stat.Size).Scan(&storageUsed)
	if errors.Is(err, pgx.ErrNoRows) {
		return chat.Message{}, apperr.QuotaExceeded
	}
	if err != nil {
		return chat.Message{}, err
	}
	m := chat.Message{
		ID:             ulid.Make().String(),
		ConversationID: conversationID,
		OwnerID:        ownerID,
		Body:           strings.TrimSpace(body),
		CreatedAt:      now,
	}
	mimeType := pending.MimeType
	if stat.ContentType != "" {
		mimeType = stat.ContentType
	}
	if _, err := tx.Exec(ctx, `
		UPDATE files
		SET status = 'READY', size_bytes = $2, mime_type = $3, upload_expires_at = NULL,
			updated_at = $4, completed_message_id = $5
		WHERE id = $1
	`, pending.ID, stat.Size, mimeType, now, m.ID); err != nil {
		return chat.Message{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messages (id, conversation_id, owner_id, body, created_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, NULL)
	`, m.ID, m.ConversationID, m.OwnerID, m.Body, m.CreatedAt); err != nil {
		return chat.Message{}, err
	}
	fileID := pending.ID
	a := chat.Attachment{
		ID: ulid.Make().String(), MessageID: m.ID, FileID: &fileID,
		OriginalName: pending.OriginalName, Name: pending.Name, MimeType: mimeType,
		SizeBytes: stat.Size, CreatedAt: now, Availability: chat.AttachmentAvailable,
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO message_attachments
			(id, message_id, file_id, original_name, display_name, mime_type, size_bytes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, a.ID, a.MessageID, a.FileID, a.OriginalName, a.Name, a.MimeType, a.SizeBytes, a.CreatedAt); err != nil {
		return chat.Message{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE conversations SET updated_at = $3 WHERE id = $1 AND owner_id = $2
	`, conversationID, ownerID, now); err != nil {
		return chat.Message{}, err
	}
	m.Attachments = []chat.Attachment{a}
	if err := tx.Commit(ctx); err != nil {
		return chat.Message{}, err
	}
	return m, nil
}

func (s *Store) CreateAttachment(ctx context.Context, a chat.Attachment) error {
	if a.FileID == nil {
		return apperr.Validation
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO message_attachments
			(id, message_id, file_id, original_name, display_name, mime_type, size_bytes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, a.ID, a.MessageID, a.FileID, a.OriginalName, a.Name, a.MimeType, a.SizeBytes, a.CreatedAt)
	return err
}

func (s *Store) ListMessagesBefore(ctx context.Context, ownerID, conversationID, before string, limit int) ([]chat.Message, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, conversation_id, owner_id, body, created_at
		FROM (
			SELECT id, conversation_id, owner_id, body, created_at
			FROM messages
			WHERE owner_id = $1
				AND conversation_id = $2
				AND deleted_at IS NULL
				AND ($3::char(26) IS NULL OR id < $3::char(26))
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		) latest
		ORDER BY created_at ASC, id ASC
	`, ownerID, conversationID, nullableCursor(before), limit)
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

func (s *Store) LastMessageForConversation(ctx context.Context, ownerID, conversationID string) (*chat.Message, error) {
	var m chat.Message
	err := s.pool.QueryRow(ctx, `
		SELECT id, conversation_id, owner_id, body, created_at
		FROM messages
		WHERE owner_id = $1 AND conversation_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, ownerID, conversationID).Scan(&m.ID, &m.ConversationID, &m.OwnerID, &m.Body, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m.Attachments, err = s.listMessageAttachments(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func nullableCursor(before string) any {
	if before == "" {
		return nil
	}
	return before
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

func (s *Store) SearchMessages(ctx context.Context, ownerID, conversationID, query string, limit int) ([]chat.Message, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, conversation_id, owner_id, body, created_at
		FROM messages
		WHERE owner_id = $1
			AND conversation_id = $2
			AND deleted_at IS NULL
			AND body ILIKE '%' || $3 || '%'
		ORDER BY created_at DESC
		LIMIT $4
	`, ownerID, conversationID, query, limit)
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

func (s *Store) ListMedia(ctx context.Context, ownerID, conversationID string, limit int) ([]chat.Attachment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.message_id, a.file_id, a.original_name, a.display_name, a.mime_type, a.size_bytes, a.created_at,
			'available'
		FROM message_attachments a
		JOIN messages m ON m.id = a.message_id
		JOIN files f ON f.id = a.file_id
		WHERE m.owner_id = $1
			AND m.conversation_id = $2
			AND m.deleted_at IS NULL
			AND f.deleted_at IS NULL
			AND f.status = 'READY'
			AND (f.mime_type LIKE 'image/%' OR f.mime_type LIKE 'video/%')
		ORDER BY a.created_at DESC
		LIMIT $3
	`, ownerID, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []chat.Attachment
	for rows.Next() {
		var a chat.Attachment
		if err := rows.Scan(&a.ID, &a.MessageID, &a.FileID, &a.OriginalName, &a.Name, &a.MimeType, &a.SizeBytes, &a.CreatedAt, &a.Availability); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) listMessageAttachments(ctx context.Context, messageID string) ([]chat.Attachment, error) {
	return s.listMessageAttachmentsFrom(ctx, s.pool, messageID)
}

func (s *Store) listMessageAttachmentsFrom(ctx context.Context, queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, messageID string) ([]chat.Attachment, error) {
	rows, err := queryer.Query(ctx, `
		SELECT a.id, a.message_id, a.file_id, a.original_name, a.display_name, a.mime_type, a.size_bytes, a.created_at,
			CASE
				WHEN a.file_id IS NULL THEN 'purged'
				WHEN f.deleted_at IS NOT NULL THEN 'trashed'
				WHEN f.status = 'READY' THEN 'available'
				ELSE 'purged'
			END
		FROM message_attachments a
		LEFT JOIN files f ON f.id = a.file_id
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
		if err := rows.Scan(&a.ID, &a.MessageID, &a.FileID, &a.OriginalName, &a.Name, &a.MimeType, &a.SizeBytes, &a.CreatedAt, &a.Availability); err != nil {
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

func (r chatRepo) CreateMessageWithAttachments(ctx context.Context, m chat.Message, fileIDs []string) (chat.Message, error) {
	return r.store.CreateMessageWithAttachments(ctx, m, fileIDs)
}

func (r chatRepo) ListMessages(ctx context.Context, ownerID, conversationID string, limit int) ([]chat.Message, error) {
	return r.store.ListMessages(ctx, ownerID, conversationID, limit)
}

func (r chatRepo) ListMessagesBefore(ctx context.Context, ownerID, conversationID, before string, limit int) ([]chat.Message, error) {
	return r.store.ListMessagesBefore(ctx, ownerID, conversationID, before, limit)
}

func (r chatRepo) LastMessageForConversation(ctx context.Context, ownerID, conversationID string) (*chat.Message, error) {
	return r.store.LastMessageForConversation(ctx, ownerID, conversationID)
}

func (r chatRepo) SearchMessages(ctx context.Context, ownerID, conversationID, query string, limit int) ([]chat.Message, error) {
	return r.store.SearchMessages(ctx, ownerID, conversationID, query, limit)
}

func (r chatRepo) CreateAttachment(ctx context.Context, a chat.Attachment) error {
	return r.store.CreateAttachment(ctx, a)
}

func (r chatRepo) ListMedia(ctx context.Context, ownerID, conversationID string, limit int) ([]chat.Attachment, error) {
	return r.store.ListMedia(ctx, ownerID, conversationID, limit)
}

func (r chatRepo) CompleteAttachment(ctx context.Context, ownerID, conversationID string, f file.File, stat objectstore.ObjectStat, body string, now time.Time) (chat.Message, error) {
	return r.store.CompleteAttachment(ctx, ownerID, conversationID, f, stat, body, now)
}

var _ chat.Repository = chatRepo{}
