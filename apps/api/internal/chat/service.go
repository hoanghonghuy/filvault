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
	ListMessages(ctx context.Context, ownerID, conversationID string, limit int) ([]Message, error)
	SearchMessages(ctx context.Context, ownerID, conversationID, query string, limit int) ([]Message, error)
	CreateAttachment(ctx context.Context, a Attachment) error
	ListMedia(ctx context.Context, ownerID, conversationID string, limit int) ([]Attachment, error)
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
	if err := s.repo.CreateMessage(ctx, m); err != nil {
		return Message{}, err
	}
	for _, fileID := range fileIDs {
		f, err := s.files.GetByID(ctx, ownerID, fileID)
		if err != nil {
			return Message{}, err
		}
		if f == nil || f.Status != file.StatusReady {
			return Message{}, apperr.NotFound
		}
		a := Attachment{
			ID:           ulid.Make().String(),
			MessageID:    m.ID,
			FileID:       f.ID,
			OriginalName: f.OriginalName,
			Name:         f.Name,
			MimeType:     f.MimeType,
			SizeBytes:    f.SizeBytes,
			CreatedAt:    now,
		}
		if err := s.repo.CreateAttachment(ctx, a); err != nil {
			return Message{}, err
		}
		m.Attachments = append(m.Attachments, a)
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
	return s.repo.ListMedia(ctx, ownerID, conversationID, limit)
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
	if f == nil || f.Status != file.StatusPending {
		return Message{}, apperr.NotFound
	}
	now := s.now().UTC()
	if f.UploadExpiresAt != nil && !now.Before(*f.UploadExpiresAt) {
		return Message{}, apperr.UploadExpired
	}
	stat, err := s.objects.Head(ctx, f.ObjectKey)
	if err != nil {
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			return Message{}, apperr.UploadExpired
		}
		return Message{}, err
	}
	used, quota, err := s.quota.GetStorage(ctx, ownerID)
	if err != nil {
		return Message{}, err
	}
	if used+stat.Size > quota {
		return Message{}, apperr.QuotaExceeded
	}
	ready := *f
	ready.Status = file.StatusReady
	ready.SizeBytes = stat.Size
	if stat.ContentType != "" {
		ready.MimeType = stat.ContentType
	}
	ready.UploadExpiresAt = nil
	ready.UpdatedAt = now
	if err := s.files.Update(ctx, ready); err != nil {
		return Message{}, err
	}
	if err := s.quota.AddStorageUsed(ctx, ownerID, stat.Size); err != nil {
		return Message{}, err
	}
	return s.CreateMessage(ctx, ownerID, conversationID, body, []string{fileID})
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
