package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/chat"
	"filvault/internal/file"
	"filvault/internal/platform/objectstore"

	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

func (s *Store) CreateConversation(ctx context.Context, c chat.Conversation) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		INSERT INTO conversations
			(id, owner_id, title, created_at, updated_at, archived_at,
			 conversation_type, created_by_user_id, last_message_at)
		VALUES ($1, $2, $3, $4, $5, NULL, 'legacy', $2, $5)
	`, c.ID, c.OwnerID, c.Title, c.CreatedAt, c.UpdatedAt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO conversation_members (conversation_id, user_id, joined_at)
		VALUES ($1, $2, $3)
	`, c.ID, c.OwnerID, c.CreatedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CreateOrGetDirectConversation(ctx context.Context, ownerID, recipientID string) (chat.Conversation, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return chat.Conversation{}, err
	}
	defer tx.Rollback(ctx)

	ids := []string{ownerID, recipientID}
	if ids[0] > ids[1] {
		ids[0], ids[1] = ids[1], ids[0]
	}
	directKey := ids[0] + ":" + ids[1]
	var c chat.Conversation
	err = tx.QueryRow(ctx, `
		SELECT id, owner_id, COALESCE(title, ''), conversation_type,
			created_at, updated_at, COALESCE(last_message_at, updated_at)
		FROM conversations
		WHERE direct_key = $1
		FOR UPDATE
	`, directKey).Scan(
		&c.ID, &c.OwnerID, &c.Title, &c.Type,
		&c.CreatedAt, &c.UpdatedAt, &c.LastMessage,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		now := time.Now().UTC()
		c = chat.Conversation{
			ID: auth.NewID(), OwnerID: ownerID, Type: "direct",
			CreatedAt: now, UpdatedAt: now, LastMessage: now,
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO conversations
				(id, owner_id, title, created_at, updated_at, archived_at,
				 conversation_type, direct_key, created_by_user_id, last_message_at)
			VALUES ($1, $2, NULL, $3, $3, NULL, 'direct', $4, $2, $3)
		`, c.ID, ownerID, now, directKey); err != nil {
			if isUniqueViolation(err) {
				_ = tx.Rollback(ctx)
				return s.CreateOrGetDirectConversation(ctx, ownerID, recipientID)
			}
			return chat.Conversation{}, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO conversation_members (conversation_id, user_id, joined_at)
			VALUES ($1, $2, $3), ($1, $4, $3)
		`, c.ID, ownerID, now, recipientID); err != nil {
			return chat.Conversation{}, err
		}
	} else if err != nil {
		return chat.Conversation{}, err
	} else {
		if _, err := tx.Exec(ctx, `
			INSERT INTO conversation_members (conversation_id, user_id, joined_at)
			VALUES ($1, $2, $3), ($1, $4, $3)
			ON CONFLICT (conversation_id, user_id) DO UPDATE SET archived_at = NULL
		`, c.ID, ownerID, c.CreatedAt, recipientID); err != nil {
			return chat.Conversation{}, err
		}
	}
	_ = tx.QueryRow(ctx, `
		SELECT u.id, u.display_name, u.email
		FROM conversation_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.conversation_id = $1 AND cm.user_id <> $2
	`, c.ID, ownerID).Scan(&c.PeerID, &c.PeerName, &c.PeerEmail)
	if err := tx.Commit(ctx); err != nil {
		return chat.Conversation{}, err
	}
	return c, nil
}

func (s *Store) ListConversations(ctx context.Context, ownerID string) ([]chat.Conversation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.owner_id, COALESCE(c.title, ''), c.created_at, c.updated_at,
			COALESCE(c.conversation_type, 'legacy'),
			COALESCE(peer.id, ''), COALESCE(peer.display_name, ''), COALESCE(peer.email, ''),
			COALESCE(c.last_message_at, c.updated_at)
		FROM conversations c
		LEFT JOIN conversation_members cm ON cm.conversation_id = c.id AND cm.user_id <> $1
		LEFT JOIN users peer ON peer.id = cm.user_id
		WHERE c.archived_at IS NULL
			AND EXISTS (
				SELECT 1 FROM conversation_members member
				WHERE member.conversation_id = c.id AND member.user_id = $1 AND member.archived_at IS NULL
			)
		ORDER BY COALESCE(c.last_message_at, c.updated_at) DESC, c.id DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []chat.Conversation
	for rows.Next() {
		var c chat.Conversation
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Title, &c.CreatedAt, &c.UpdatedAt,
			&c.Type, &c.PeerID, &c.PeerName, &c.PeerEmail, &c.LastMessage); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetConversation(ctx context.Context, ownerID, id string) (*chat.Conversation, error) {
	var c chat.Conversation
	err := s.pool.QueryRow(ctx, `
		SELECT c.id, c.owner_id, COALESCE(c.title, ''), c.created_at, c.updated_at,
			COALESCE(c.conversation_type, 'legacy'),
			COALESCE(peer.id, ''), COALESCE(peer.display_name, ''), COALESCE(peer.email, ''),
			COALESCE(c.last_message_at, c.updated_at)
		FROM conversations c
		LEFT JOIN conversation_members cm ON cm.conversation_id = c.id AND cm.user_id <> $1
		LEFT JOIN users peer ON peer.id = cm.user_id
		WHERE c.id = $2 AND c.archived_at IS NULL
			AND EXISTS (
				SELECT 1 FROM conversation_members member
				WHERE member.conversation_id = c.id AND member.user_id = $1 AND member.archived_at IS NULL
			)
	`, ownerID, id).Scan(
		&c.ID, &c.OwnerID, &c.Title, &c.CreatedAt, &c.UpdatedAt,
		&c.Type, &c.PeerID, &c.PeerName, &c.PeerEmail, &c.LastMessage,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) CreateMessage(ctx context.Context, m chat.Message) (chat.Message, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return chat.Message{}, err
	}
	defer tx.Rollback(ctx)
	if m.ClientMessageID != "" {
		var existing chat.Message
		err = tx.QueryRow(ctx, `
			SELECT id, conversation_id, owner_id, sender_id, COALESCE(client_message_id, ''),
				body, created_at, edited_at, removed_at
			FROM messages
			WHERE conversation_id = $1 AND sender_id = $2 AND client_message_id = $3
		`, m.ConversationID, m.SenderID, m.ClientMessageID).Scan(
			&existing.ID, &existing.ConversationID, &existing.OwnerID, &existing.SenderID,
			&existing.ClientMessageID, &existing.Body, &existing.CreatedAt,
			&existing.EditedAt, &existing.RemovedAt,
		)
		if err == nil {
			if existing.Body != m.Body {
				return chat.Message{}, apperr.Conflict
			}
			if err := tx.Commit(ctx); err != nil {
				return chat.Message{}, err
			}
			existing.Attachments, err = s.listMessageAttachments(ctx, existing.ID)
			existing.Idempotent = true
			return existing, err
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return chat.Message{}, err
		}
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO messages
			(id, conversation_id, owner_id, sender_id, client_message_id, body, created_at, deleted_at)
		VALUES ($1, $2, $3, $3, NULLIF($4::text, ''), $5, $6, NULL)
	`, m.ID, m.ConversationID, m.OwnerID, m.ClientMessageID, m.Body, m.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) && m.ClientMessageID != "" {
			_ = tx.Rollback(ctx)
			return s.CreateMessage(ctx, m)
		}
		return chat.Message{}, err
	}
	if _, err = tx.Exec(ctx, `
		UPDATE conversations SET updated_at = $2, last_message_at = $2
		WHERE id = $1
	`, m.ConversationID, m.CreatedAt); err != nil {
		return chat.Message{}, err
	}
	if err := insertChatEvent(ctx, tx, m.ConversationID, "message.created", m.ID, m.CreatedAt); err != nil {
		return chat.Message{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return chat.Message{}, err
	}
	return m, nil
}

func (s *Store) UpdateMessage(ctx context.Context, userID, conversationID, messageID, body string, now time.Time) (chat.Message, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return chat.Message{}, err
	}
	defer tx.Rollback(ctx)
	var m chat.Message
	err = tx.QueryRow(ctx, `
		UPDATE messages
		SET body = $4, edited_at = $5
		WHERE id = $1 AND conversation_id = $2 AND sender_id = $3
			AND removed_at IS NULL AND created_at >= $5::timestamptz - INTERVAL '15 minutes'
		RETURNING id, conversation_id, owner_id, sender_id, client_message_id,
			body, created_at, edited_at, removed_at
	`, messageID, conversationID, userID, body, now).Scan(
		&m.ID, &m.ConversationID, &m.OwnerID, &m.SenderID, &m.ClientMessageID,
		&m.Body, &m.CreatedAt, &m.EditedAt, &m.RemovedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return chat.Message{}, apperr.Forbidden
	}
	if err != nil {
		return chat.Message{}, err
	}
	if err := insertChatEvent(ctx, tx, conversationID, "message.edited", m.ID, now); err != nil {
		return chat.Message{}, err
	}
	m.Attachments, err = s.listMessageAttachmentsFrom(ctx, tx, m.ID)
	if err != nil {
		return chat.Message{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return chat.Message{}, err
	}
	return m, err
}

func (s *Store) RemoveMessage(ctx context.Context, userID, conversationID, messageID string, now time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `
		UPDATE messages
		SET removed_at = $4, removed_by_user_id = $3, body = ''
		WHERE id = $1 AND conversation_id = $2 AND sender_id = $3
			AND removed_at IS NULL AND created_at >= $4::timestamptz - INTERVAL '15 minutes'
	`, messageID, conversationID, userID, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.Forbidden
	}
	if _, err := tx.Exec(ctx, `
		UPDATE chat_shared_album_items
		SET hidden_at = $2
		WHERE attachment_id IN (
			SELECT id FROM message_attachments WHERE message_id = $1
		) AND hidden_at IS NULL
	`, messageID, now); err != nil {
		return err
	}
	if err := insertChatEvent(ctx, tx, conversationID, "message.removed", messageID, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CreateMessageWithAttachments(ctx context.Context, m chat.Message, fileIDs []string) (chat.Message, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return chat.Message{}, err
	}
	defer tx.Rollback(ctx)

	if m.ClientMessageID != "" {
		var existing chat.Message
		err = tx.QueryRow(ctx, `
			SELECT id, conversation_id, owner_id, sender_id, COALESCE(client_message_id, ''),
				body, created_at, edited_at, removed_at
			FROM messages
			WHERE conversation_id = $1 AND sender_id = $2 AND client_message_id = $3
		`, m.ConversationID, m.SenderID, m.ClientMessageID).Scan(
			&existing.ID, &existing.ConversationID, &existing.OwnerID, &existing.SenderID,
			&existing.ClientMessageID, &existing.Body, &existing.CreatedAt,
			&existing.EditedAt, &existing.RemovedAt,
		)
		if err == nil {
			if existing.Body != m.Body {
				return chat.Message{}, apperr.Conflict
			}
			existing.Attachments, err = s.listMessageAttachmentsFrom(ctx, tx, existing.ID)
			if err != nil {
				return chat.Message{}, err
			}
			existing.Idempotent = true
			if err := tx.Commit(ctx); err != nil {
				return chat.Message{}, err
			}
			return existing, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return chat.Message{}, err
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO messages
			(id, conversation_id, owner_id, sender_id, client_message_id, body, created_at, deleted_at)
		VALUES ($1, $2, $3, $3, NULLIF($4::text, ''), $5, $6, NULL)
	`, m.ID, m.ConversationID, m.OwnerID, m.ClientMessageID, m.Body, m.CreatedAt); err != nil {
		if isUniqueViolation(err) && m.ClientMessageID != "" {
			_ = tx.Rollback(ctx)
			return s.CreateMessageWithAttachments(ctx, m, fileIDs)
		}
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
		albumID, err := ensureChatSharedAlbum(ctx, tx, m.ConversationID, m.CreatedAt)
		if err != nil {
			return chat.Message{}, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO chat_shared_album_items (album_id, attachment_id, added_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (attachment_id) DO NOTHING
		`, albumID, a.ID, a.CreatedAt); err != nil {
			return chat.Message{}, err
		}
		m.Attachments = append(m.Attachments, a)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE conversations SET updated_at = $2, last_message_at = $2
		WHERE id = $1
		`, m.ConversationID, m.CreatedAt); err != nil {
		return chat.Message{}, err
	}
	if err := insertChatEvent(ctx, tx, m.ConversationID, "message.created", m.ID, m.CreatedAt); err != nil {
		return chat.Message{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return chat.Message{}, err
	}
	return m, nil
}

func ensureChatSharedAlbum(ctx context.Context, tx pgx.Tx, conversationID string, now time.Time) (string, error) {
	albumID := auth.NewID()
	err := tx.QueryRow(ctx, `
		INSERT INTO chat_shared_albums (id, conversation_id, name, created_at, updated_at)
		VALUES ($1, $2, 'Shared media', $3, $3)
		ON CONFLICT (conversation_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
		RETURNING id
	`, albumID, conversationID, now).Scan(&albumID)
	return albumID, err
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
			SELECT id, conversation_id, owner_id, sender_id, COALESCE(client_message_id, ''),
				body, created_at, edited_at, removed_at
			FROM messages
			WHERE id = $1 AND owner_id = $2
		`, *completedMessageID, ownerID).Scan(
			&m.ID, &m.ConversationID, &m.OwnerID, &m.SenderID, &m.ClientMessageID,
			&m.Body, &m.CreatedAt, &m.EditedAt, &m.RemovedAt,
		)
		if err != nil {
			return chat.Message{}, err
		}
		if m.ConversationID != conversationID {
			return chat.Message{}, apperr.NotFound
		}
		m.Idempotent = true
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
		SELECT owner_id FROM conversations
		WHERE id = $1 AND archived_at IS NULL
			AND EXISTS (
				SELECT 1 FROM conversation_members
				WHERE conversation_id = $1 AND user_id = $2 AND archived_at IS NULL
			)
		FOR UPDATE
	`, conversationID, ownerID).Scan(&conversationOwner); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return chat.Message{}, apperr.NotFound
		}
		return chat.Message{}, err
	}
	m := chat.Message{
		ID:             ulid.Make().String(),
		ConversationID: conversationID,
		OwnerID:        ownerID,
		SenderID:       ownerID,
		Body:           strings.TrimSpace(body),
		CreatedAt:      now,
	}
	mimeType := pending.MimeType
	if stat.ContentType != "" {
		mimeType = stat.ContentType
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
	if _, err := tx.Exec(ctx, `
		UPDATE files
		SET status = 'READY', size_bytes = $2, mime_type = $3, upload_expires_at = NULL,
			updated_at = $4, completed_message_id = $5
		WHERE id = $1
	`, pending.ID, stat.Size, mimeType, now, m.ID); err != nil {
		return chat.Message{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messages
			(id, conversation_id, owner_id, sender_id, client_message_id, body, created_at, deleted_at)
		VALUES ($1, $2, $3, $3, NULLIF($4::text, ''), $5, $6, NULL)
	`, m.ID, m.ConversationID, m.OwnerID, m.ClientMessageID, m.Body, m.CreatedAt); err != nil {
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
	albumID, err := ensureChatSharedAlbum(ctx, tx, conversationID, now)
	if err != nil {
		return chat.Message{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO chat_shared_album_items (album_id, attachment_id, added_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (attachment_id) DO UPDATE SET hidden_at = NULL
	`, albumID, a.ID, a.CreatedAt); err != nil {
		return chat.Message{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE conversations SET updated_at = $2, last_message_at = $2 WHERE id = $1
	`, conversationID, now); err != nil {
		return chat.Message{}, err
	}
	if err := insertChatEvent(ctx, tx, conversationID, "message.created", m.ID, now); err != nil {
		return chat.Message{}, err
	}
	m.Attachments = []chat.Attachment{a}
	if err := tx.Commit(ctx); err != nil {
		return chat.Message{}, err
	}
	return m, nil
}

func insertChatEvent(ctx context.Context, tx pgx.Tx, conversationID, eventType, aggregateID string, now time.Time) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO chat_events (id, conversation_id, event_type, aggregate_id, payload, created_at)
		VALUES ($1, $2, $3, $4::char(26), jsonb_build_object('aggregateId', $4::text), $5)
	`, auth.NewID(), conversationID, eventType, aggregateID, now)
	return err
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
		SELECT latest.id, latest.conversation_id, latest.owner_id, latest.sender_id,
			latest.client_message_id, latest.body, latest.created_at, latest.edited_at, latest.removed_at
		FROM (
			SELECT id, conversation_id, owner_id, sender_id, COALESCE(client_message_id, '') AS client_message_id, body, created_at, edited_at, removed_at
			FROM messages
			WHERE conversation_id = $2
				AND EXISTS (
					SELECT 1 FROM conversation_members
					WHERE conversation_id = $2 AND user_id = $1 AND archived_at IS NULL
				)
				AND deleted_at IS NULL
				AND (
					$3::char(26) IS NULL OR
					(created_at, id) < (
						SELECT created_at, id
						FROM messages
						WHERE id = $3::char(26) AND conversation_id = $2
					)
				)
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		) latest
		ORDER BY latest.created_at ASC, latest.id ASC
	`, ownerID, conversationID, nullableCursor(before), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []chat.Message
	for rows.Next() {
		var m chat.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.OwnerID, &m.SenderID, &m.ClientMessageID, &m.Body, &m.CreatedAt, &m.EditedAt, &m.RemovedAt); err != nil {
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
			SELECT id, conversation_id, owner_id, sender_id, COALESCE(client_message_id, ''), body, created_at, edited_at, removed_at
		FROM messages
		WHERE conversation_id = $2
			AND EXISTS (
				SELECT 1 FROM conversation_members
				WHERE conversation_id = $2 AND user_id = $1 AND archived_at IS NULL
			)
			AND deleted_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, ownerID, conversationID).Scan(&m.ID, &m.ConversationID, &m.OwnerID, &m.SenderID, &m.ClientMessageID, &m.Body, &m.CreatedAt, &m.EditedAt, &m.RemovedAt)
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
		SELECT id, conversation_id, owner_id, sender_id, COALESCE(client_message_id, ''), body, created_at, edited_at, removed_at
		FROM messages
		WHERE conversation_id = $2
			AND EXISTS (
				SELECT 1 FROM conversation_members
				WHERE conversation_id = $2 AND user_id = $1 AND archived_at IS NULL
			)
			AND deleted_at IS NULL
		ORDER BY created_at ASC, id ASC
		LIMIT $3
	`, ownerID, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []chat.Message
	for rows.Next() {
		var m chat.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.OwnerID, &m.SenderID, &m.ClientMessageID, &m.Body, &m.CreatedAt, &m.EditedAt, &m.RemovedAt); err != nil {
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
		SELECT id, conversation_id, owner_id, sender_id, COALESCE(client_message_id, ''), body, created_at, edited_at, removed_at
		FROM messages
		WHERE conversation_id = $2
			AND EXISTS (
				SELECT 1 FROM conversation_members
				WHERE conversation_id = $2 AND user_id = $1 AND archived_at IS NULL
			)
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
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.OwnerID, &m.SenderID, &m.ClientMessageID, &m.Body, &m.CreatedAt, &m.EditedAt, &m.RemovedAt); err != nil {
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
			CASE
				WHEN a.file_id IS NULL THEN 'purged'
				WHEN f.deleted_at IS NOT NULL THEN 'trashed'
				WHEN f.status = 'READY' THEN 'available'
				ELSE 'purged'
			END
		FROM message_attachments a
		JOIN messages m ON m.id = a.message_id
		JOIN chat_shared_album_items sai ON sai.attachment_id = a.id AND sai.hidden_at IS NULL
		LEFT JOIN files f ON f.id = a.file_id
		WHERE m.conversation_id = $2
			AND EXISTS (
				SELECT 1 FROM conversation_members
				WHERE conversation_id = $2 AND user_id = $1 AND archived_at IS NULL
			)
			AND m.deleted_at IS NULL
			AND (a.mime_type LIKE 'image/%' OR a.mime_type LIKE 'video/%')
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

func (s *Store) GetAttachmentObjectKey(ctx context.Context, userID, conversationID, attachmentID string) (string, error) {
	var key string
	err := s.pool.QueryRow(ctx, `
		SELECT f.object_key
		FROM message_attachments a
		JOIN messages m ON m.id = a.message_id
		JOIN files f ON f.id = a.file_id
		WHERE a.id = $1
			AND m.conversation_id = $2
			AND f.deleted_at IS NULL
			AND f.status = 'READY'
			AND EXISTS (
				SELECT 1 FROM conversation_members
				WHERE conversation_id = $2 AND user_id = $3 AND archived_at IS NULL
			)
	`, attachmentID, conversationID, userID).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", apperr.NotFound
	}
	return key, err
}

func (s *Store) ListChatEventsAfter(ctx context.Context, userID string, after int64, limit int) ([]chat.Event, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.sequence, e.conversation_id, e.event_type, e.aggregate_id, e.payload, e.created_at
		FROM chat_events e
		WHERE e.sequence > $1
			AND EXISTS (
				SELECT 1 FROM conversation_members cm
				WHERE cm.conversation_id = e.conversation_id
					AND cm.user_id = $2 AND cm.archived_at IS NULL
			)
		ORDER BY e.sequence ASC
		LIMIT $3
	`, after, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []chat.Event
	for rows.Next() {
		var event chat.Event
		if err := rows.Scan(&event.ID, &event.Sequence, &event.ConversationID, &event.Type, &event.AggregateID, &event.Payload, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
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

func (r chatRepo) CreateOrGetDirectConversation(ctx context.Context, ownerID, recipientID string) (chat.Conversation, error) {
	return r.store.CreateOrGetDirectConversation(ctx, ownerID, recipientID)
}

func (r chatRepo) ListConversations(ctx context.Context, ownerID string) ([]chat.Conversation, error) {
	return r.store.ListConversations(ctx, ownerID)
}

func (r chatRepo) GetConversation(ctx context.Context, ownerID, id string) (*chat.Conversation, error) {
	return r.store.GetConversation(ctx, ownerID, id)
}

func (r chatRepo) CreateMessage(ctx context.Context, m chat.Message) (chat.Message, error) {
	return r.store.CreateMessage(ctx, m)
}

func (r chatRepo) UpdateMessage(ctx context.Context, userID, conversationID, messageID, body string, now time.Time) (chat.Message, error) {
	return r.store.UpdateMessage(ctx, userID, conversationID, messageID, body, now)
}

func (r chatRepo) RemoveMessage(ctx context.Context, userID, conversationID, messageID string, now time.Time) error {
	return r.store.RemoveMessage(ctx, userID, conversationID, messageID, now)
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

func (r chatRepo) GetAttachmentObjectKey(ctx context.Context, userID, conversationID, attachmentID string) (string, error) {
	return r.store.GetAttachmentObjectKey(ctx, userID, conversationID, attachmentID)
}

func (r chatRepo) ListEventsAfter(ctx context.Context, userID string, after int64, limit int) ([]chat.Event, error) {
	return r.store.ListChatEventsAfter(ctx, userID, after, limit)
}

func (r chatRepo) CompleteAttachment(ctx context.Context, ownerID, conversationID string, f file.File, stat objectstore.ObjectStat, body string, now time.Time) (chat.Message, error) {
	return r.store.CompleteAttachment(ctx, ownerID, conversationID, f, stat, body, now)
}

func (s *Store) InsertChatEvent(ctx context.Context, conversationID, eventType, aggregateID string, payload []byte, now time.Time) error {
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO chat_events (id, conversation_id, event_type, aggregate_id, payload, created_at)
		VALUES ($1, $2, $3, $4::char(26), $5::jsonb, $6)
	`, auth.NewID(), conversationID, eventType, aggregateID, payload, now)
	return err
}

func (r chatRepo) InsertChatEvent(ctx context.Context, conversationID, eventType, aggregateID string, payload []byte, now time.Time) error {
	return r.store.InsertChatEvent(ctx, conversationID, eventType, aggregateID, payload, now)
}

var _ chat.Repository = chatRepo{}

