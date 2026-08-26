package chat

import (
	"context"
	"errors"
	"strings"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/file"
	"filvault/internal/platform/config"
	"filvault/internal/platform/objectstore"

	"github.com/oklog/ulid/v2"
)

type Repository interface {
	CreateConversation(ctx context.Context, c Conversation) error
	ListConversations(ctx context.Context, ownerID string) ([]Conversation, error)
	GetConversation(ctx context.Context, ownerID, id string) (*Conversation, error)
	CreateMessage(ctx context.Context, m Message) error
	CreateMessageWithAttachments(ctx context.Context, m Message, fileIDs []string) (Message, error)
	ListMessages(ctx context.Context, ownerID, conversationID string, limit int) ([]Message, error)
	ListMessagesBefore(ctx context.Context, ownerID, conversationID, before string, limit int) ([]Message, error)
	LastMessageForConversation(ctx context.Context, ownerID, conversationID string) (*Message, error)
	SearchMessages(ctx context.Context, ownerID, conversationID, query string, limit int) ([]Message, error)
	CreateAttachment(ctx context.Context, a Attachment) error
	ListMedia(ctx context.Context, ownerID, conversationID string, limit int) ([]Attachment, error)
	CompleteAttachment(ctx context.Context, ownerID, conversationID string, f file.File, stat objectstore.ObjectStat, body string, now time.Time) (Message, error)
}

type MessagePage struct {
	Messages   []Message
	HasMore    bool
	NextBefore string
}

type ConversationView struct {
	Conversation
	Preview *ConversationPreview `json:"preview,omitempty"`
}

type Service struct {
	repo    Repository
	files   file.Repository
	quota   file.QuotaStore
	objects objectstore.ObjectStore
	now     func() time.Time
}

func NewService(repo Repository, files file.Repository, quota file.QuotaStore, objects objectstore.ObjectStore) *Service {
	return &Service{repo: repo, files: files, quota: quota, objects: objects, now: time.Now}
}

func (s *Service) CreateConversation(ctx context.Context, ownerID, title string) (Conversation, error) {
	title = strings.TrimSpace(title)
	now := s.now().UTC()
	c := Conversation{ID: ulid.Make().String(), OwnerID: ownerID, Title: title, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateConversation(ctx, c); err != nil {
		return Conversation{}, err
	}
	return c, nil
}

func (s *Service) ListConversations(ctx context.Context, ownerID string) ([]Conversation, error) {
	return s.repo.ListConversations(ctx, ownerID)
}

func (s *Service) CreateMessage(ctx context.Context, ownerID, conversationID, body string, fileIDs []string) (Message, error) {
	body = strings.TrimSpace(body)
	if body == "" && len(fileIDs) == 0 {
		return Message{}, apperr.Validation
	}
	if err := s.ensureConversation(ctx, ownerID, conversationID); err != nil {
		return Message{}, err
	}
	now := s.now().UTC()
	m := Message{ID: ulid.Make().String(), ConversationID: conversationID, OwnerID: ownerID, Body: body, CreatedAt: now}
	if len(fileIDs) > 0 {
		return s.repo.CreateMessageWithAttachments(ctx, m, fileIDs)
	}
	if err := s.repo.CreateMessage(ctx, m); err != nil {
		return Message{}, err
	}
	return m, nil
}

func (s *Service) ListMessages(ctx context.Context, ownerID, conversationID string, limit int) ([]Message, error) {
	if err := s.ensureConversation(ctx, ownerID, conversationID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.ListMessages(ctx, ownerID, conversationID, limit)
}

func (s *Service) ListMessagePage(ctx context.Context, ownerID, conversationID, before string, limit int) (MessagePage, error) {
	if err := s.ensureConversation(ctx, ownerID, conversationID); err != nil {
		return MessagePage{}, err
	}
	if limit <= 0 || limit > 100 {
		return MessagePage{}, apperr.Validation
	}
	if before != "" {
		if _, err := ulid.ParseStrict(before); err != nil {
			return MessagePage{}, apperr.Validation
		}
	}
	messages, err := s.repo.ListMessagesBefore(ctx, ownerID, conversationID, before, limit)
	if err != nil {
		return MessagePage{}, err
	}
	for i := range messages {
		for j := range messages[i].Attachments {
			if err := s.presignThumbnail(ctx, ownerID, &messages[i].Attachments[j]); err != nil {
				return MessagePage{}, err
			}
		}
	}
	page := MessagePage{Messages: messages, HasMore: len(messages) == limit}
	if page.HasMore && len(messages) > 0 {
		page.NextBefore = messages[0].ID
	}
	return page, nil
}

func (s *Service) ListConversationsWithPreview(ctx context.Context, ownerID string) ([]ConversationView, error) {
	conversations, err := s.repo.ListConversations(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]ConversationView, 0, len(conversations))
	for _, conv := range conversations {
		view := ConversationView{Conversation: conv}
		preview, err := s.repo.LastMessageForConversation(ctx, ownerID, conv.ID)
		if err != nil {
			return nil, err
		}
		if preview != nil {
			view.Preview = &ConversationPreview{
				MessageID:     preview.ID,
				Body:          preview.Body,
				CreatedAt:     preview.CreatedAt,
				AttachmentIDs: attachmentIDs(preview.Attachments),
			}
		}
		out = append(out, view)
	}
	return out, nil
}

func attachmentIDs(attachments []Attachment) []string {
	ids := make([]string, 0, len(attachments))
	for _, a := range attachments {
		if a.FileID != nil {
			ids = append(ids, *a.FileID)
		}
	}
	return ids
}

func (s *Service) SearchMessages(ctx context.Context, ownerID, conversationID, query string, limit int) ([]Message, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 || len(query) > 100 {
		return nil, apperr.Validation
	}
	if err := s.ensureConversation(ctx, ownerID, conversationID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.SearchMessages(ctx, ownerID, conversationID, query, limit)
}

func (s *Service) ListMedia(ctx context.Context, ownerID, conversationID string, limit int) ([]Attachment, error) {
	if err := s.ensureConversation(ctx, ownerID, conversationID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	items, err := s.repo.ListMedia(ctx, ownerID, conversationID, limit)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if err := s.presignThumbnail(ctx, ownerID, &items[i]); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (s *Service) presignThumbnail(ctx context.Context, ownerID string, a *Attachment) error {
	if a.Availability != AttachmentAvailable || a.FileID == nil {
		return nil
	}
	if !strings.HasPrefix(a.MimeType, "image/") && !strings.HasPrefix(a.MimeType, "video/") {
		return nil
	}
	f, err := s.files.GetByID(ctx, ownerID, *a.FileID)
	if err != nil {
		return err
	}
	if f == nil || s.objects == nil {
		return nil
	}
	url, err := s.objects.CreateDownloadURL(ctx, f.ObjectKey, objectstore.DownloadOptions{
		Expires: config.ThumbnailPresignTTL,
	})
	if err != nil {
		return err
	}
	a.ThumbnailURL = url.URL
	return nil
}

func (s *Service) CreateAttachmentSession(ctx context.Context, ownerID, conversationID, name, contentType string, size int64) (UploadSession, error) {
	if err := s.ensureConversation(ctx, ownerID, conversationID); err != nil {
		return UploadSession{}, err
	}
	if err := file.ValidateUpload(name, contentType, size); err != nil {
		return UploadSession{}, err
	}
	now := s.now().UTC()
	fileID := ulid.Make().String()
	key := file.ObjectKey(ownerID, fileID)
	expiresAt := now.Add(config.UploadPresignTTL)
	f := file.File{
		ID:              fileID,
		OwnerID:         ownerID,
		Name:            strings.TrimSpace(name),
		OriginalName:    strings.TrimSpace(name),
		ObjectKey:       key,
		MimeType:        strings.TrimSpace(contentType),
		SizeBytes:       size,
		Status:          file.StatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
		UploadExpiresAt: &expiresAt,
		Source:          "chat",
		SourceRefID:     &conversationID,
	}
	if err := s.files.Create(ctx, f); err != nil {
		return UploadSession{}, err
	}
	presigned, err := s.objects.CreateUploadURL(ctx, key, objectstore.UploadOptions{
		ContentType: f.MimeType,
		Expires:     config.UploadPresignTTL,
	})
	if err != nil {
		_ = s.files.DeleteRow(ctx, ownerID, fileID)
		return UploadSession{}, err
	}
	return UploadSession{FileID: fileID, UploadURL: presigned.URL, ExpiresAt: presigned.ExpiresAt}, nil
}

func (s *Service) CompleteAttachment(ctx context.Context, ownerID, conversationID, fileID, body string) (Message, error) {
	if err := s.ensureConversation(ctx, ownerID, conversationID); err != nil {
		return Message{}, err
	}
	f, err := s.files.GetByID(ctx, ownerID, fileID)
	if err != nil {
		return Message{}, err
	}
	if f == nil {
		return Message{}, apperr.NotFound
	}
	now := s.now().UTC()
	if f.Status == file.StatusPending && f.UploadExpiresAt != nil && !now.Before(*f.UploadExpiresAt) {
		return Message{}, apperr.UploadExpired
	}
	stat, err := s.objects.Head(ctx, f.ObjectKey)
	if err != nil {
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			return Message{}, apperr.UploadExpired
		}
		return Message{}, err
	}
	return s.repo.CompleteAttachment(ctx, ownerID, conversationID, *f, stat, body, now)
}

func (s *Service) ensureConversation(ctx context.Context, ownerID, conversationID string) error {
	c, err := s.repo.GetConversation(ctx, ownerID, conversationID)
	if err != nil {
		return err
	}
	if c == nil {
		return apperr.NotFound
	}
	return nil
}
