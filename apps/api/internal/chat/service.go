package chat

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/file"
	"filvault/internal/platform/config"
	"filvault/internal/platform/objectstore"
	"filvault/internal/user"

	"github.com/oklog/ulid/v2"
)

type Repository interface {
	CreateConversation(ctx context.Context, c Conversation) error
	CreateOrGetDirectConversation(ctx context.Context, ownerID, recipientID string) (Conversation, error)
	ListConversations(ctx context.Context, ownerID string) ([]Conversation, error)
	GetConversation(ctx context.Context, ownerID, id string) (*Conversation, error)
	CreateMessage(ctx context.Context, m Message) (Message, error)
	UpdateMessage(ctx context.Context, userID, conversationID, messageID, body string, now time.Time) (Message, error)
	RemoveMessage(ctx context.Context, userID, conversationID, messageID string, now time.Time) error
	CreateMessageWithAttachments(ctx context.Context, m Message, fileIDs []string) (Message, error)
	ListMessages(ctx context.Context, ownerID, conversationID string, limit int) ([]Message, error)
	ListMessagesBefore(ctx context.Context, ownerID, conversationID, before string, limit int) ([]Message, error)
	LastMessageForConversation(ctx context.Context, ownerID, conversationID string) (*Message, error)
	SearchMessages(ctx context.Context, ownerID, conversationID, query string, limit int) ([]Message, error)
	CreateAttachment(ctx context.Context, a Attachment) error
	ListMedia(ctx context.Context, ownerID, conversationID string, limit int) ([]Attachment, error)
	GetAttachmentObjectKey(ctx context.Context, userID, conversationID, attachmentID string) (string, error)
	ListEventsAfter(ctx context.Context, userID string, after int64, limit int) ([]Event, error)
	CompleteAttachment(ctx context.Context, ownerID, conversationID string, f file.File, stat objectstore.ObjectStat, body string, now time.Time) (Message, error)
	InsertChatEvent(ctx context.Context, conversationID, eventType, aggregateID string, payload []byte, now time.Time) error
	MarkAsRead(ctx context.Context, userID, conversationID, messageID string, now time.Time) error
	ListConversationReadStates(ctx context.Context, userID string) (map[string]ReadState, error)
}

type UserDirectory interface {
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	GetUserByID(ctx context.Context, id string) (*user.User, error)
}

type TokenGenerator interface {
	GenerateToken(room, identity, name string) (string, error)
}

type CallTokenResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
	Room  string `json:"room"`
}

type CallSignal struct {
	ConversationID string `json:"conversationId"`
	SenderID       string `json:"senderId"`
	SenderName     string `json:"senderName"`
	Action         string `json:"action"`
	IsVideo        bool   `json:"isVideo"`
	Timestamp      string `json:"timestamp"`
}

type ReadSignal struct {
	ConversationID    string `json:"conversationId"`
	UserID            string `json:"userId"`
	LastReadMessageID string `json:"lastReadMessageId,omitempty"`
	LastReadAt        string `json:"lastReadAt"`
}

type TypingSignal struct {
	ConversationID string `json:"conversationId"`
	UserID         string `json:"userId"`
	UserName       string `json:"userName"`
	Typing         bool   `json:"typing"`
}

type MessagePage struct {
	Messages   []Message
	HasMore    bool
	NextBefore string
}

type ConversationView struct {
	Conversation
	Preview               *ConversationPreview `json:"preview,omitempty"`
	UnreadCount           int                  `json:"unreadCount"`
	LastReadAt            *time.Time           `json:"lastReadAt,omitempty"`
	LastReadMessageID     string               `json:"lastReadMessageId,omitempty"`
	PeerLastReadAt        *time.Time           `json:"peerLastReadAt,omitempty"`
	PeerLastReadMessageID string               `json:"peerLastReadMessageId,omitempty"`
}

type Service struct {
	repo       Repository
	users      UserDirectory
	files      file.Repository
	quota      file.QuotaStore
	objects    objectstore.ObjectStore
	tokens     TokenGenerator
	callPubURL string
	now        func() time.Time
}

func NewService(repo Repository, users UserDirectory, files file.Repository, quota file.QuotaStore, objects objectstore.ObjectStore) *Service {
	return &Service{repo: repo, users: users, files: files, quota: quota, objects: objects, now: time.Now}
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

func (s *Service) CreateDirectConversation(ctx context.Context, ownerID, recipientEmail string) (Conversation, error) {
	recipientEmail = strings.TrimSpace(strings.ToLower(recipientEmail))
	if recipientEmail == "" || !strings.Contains(recipientEmail, "@") || s.users == nil {
		return Conversation{}, apperr.Validation
	}
	recipient, err := s.users.GetUserByEmail(ctx, recipientEmail)
	if err != nil {
		return Conversation{}, err
	}
	if recipient == nil || !recipient.EmailVerified() || recipient.ID == ownerID {
		return Conversation{}, apperr.NotFound
	}
	return s.repo.CreateOrGetDirectConversation(ctx, ownerID, recipient.ID)
}

func (s *Service) EditMessage(ctx context.Context, userID, conversationID, messageID, body string) (Message, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return Message{}, apperr.Validation
	}
	if err := s.ensureConversation(ctx, userID, conversationID); err != nil {
		return Message{}, err
	}
	return s.repo.UpdateMessage(ctx, userID, conversationID, messageID, body, s.now().UTC())
}

func (s *Service) RemoveMessage(ctx context.Context, userID, conversationID, messageID string) error {
	if err := s.ensureConversation(ctx, userID, conversationID); err != nil {
		return err
	}
	return s.repo.RemoveMessage(ctx, userID, conversationID, messageID, s.now().UTC())
}

func (s *Service) DownloadAttachment(ctx context.Context, userID, conversationID, attachmentID string) (objectstore.PresignedURL, error) {
	key, err := s.repo.GetAttachmentObjectKey(ctx, userID, conversationID, attachmentID)
	if err != nil {
		return objectstore.PresignedURL{}, err
	}
	if s.objects == nil {
		return objectstore.PresignedURL{}, apperr.InvalidState
	}
	return s.objects.CreateDownloadURL(ctx, key, objectstore.DownloadOptions{
		Expires: config.DownloadPresignTTL,
	})
}

func (s *Service) CreateMessage(ctx context.Context, ownerID, conversationID, body string, fileIDs []string, clientMessageID string) (Message, error) {
	body = strings.TrimSpace(body)
	if body == "" && len(fileIDs) == 0 {
		return Message{}, apperr.Validation
	}
	if err := s.ensureConversation(ctx, ownerID, conversationID); err != nil {
		return Message{}, err
	}
	now := s.now().UTC()
	m := Message{
		ID: ulid.Make().String(), ConversationID: conversationID, OwnerID: ownerID,
		SenderID: ownerID, Body: body, ClientMessageID: strings.TrimSpace(clientMessageID), CreatedAt: now,
	}
	if len(fileIDs) > 0 {
		return s.repo.CreateMessageWithAttachments(ctx, m, fileIDs)
	}
	created, err := s.repo.CreateMessage(ctx, m)
	if err != nil {
		return Message{}, err
	}
	return created, nil
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
	messages, err := s.repo.ListMessagesBefore(ctx, ownerID, conversationID, before, limit+1)
	if err != nil {
		return MessagePage{}, err
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[len(messages)-limit:]
	}
	for i := range messages {
		for j := range messages[i].Attachments {
			if err := s.presignThumbnail(ctx, ownerID, &messages[i].Attachments[j]); err != nil {
				return MessagePage{}, err
			}
		}
	}
	page := MessagePage{Messages: messages, HasMore: hasMore}
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
	readStates, _ := s.repo.ListConversationReadStates(ctx, ownerID)
	out := make([]ConversationView, 0, len(conversations))
	for _, conv := range conversations {
		view := ConversationView{Conversation: conv}
		if readStates != nil {
			if rs, ok := readStates[conv.ID]; ok {
				view.UnreadCount = rs.UnreadCount
				view.LastReadAt = rs.LastReadAt
				view.LastReadMessageID = rs.LastReadMessageID
				view.PeerLastReadAt = rs.PeerLastReadAt
				view.PeerLastReadMessageID = rs.PeerLastReadMessageID
			}
		}
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
	if f.Source != "chat" || f.SourceRefID == nil || *f.SourceRefID != conversationID {
		return Message{}, apperr.NotFound
	}
	now := s.now().UTC()
	if f.Status == file.StatusPending && f.UploadExpiresAt != nil && !now.Before(*f.UploadExpiresAt) {
		return Message{}, apperr.UploadExpired
	}
	var stat objectstore.ObjectStat
	if f.Status == file.StatusPending {
		if s.objects == nil {
			return Message{}, apperr.InvalidState
		}
		stat, err = s.objects.Head(ctx, f.ObjectKey)
		if err != nil {
			if errors.Is(err, objectstore.ErrObjectNotFound) {
				return Message{}, apperr.UploadExpired
			}
			return Message{}, err
		}
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

func (s *Service) SetCallConfig(tokens TokenGenerator, publicURL string) {
	s.tokens = tokens
	s.callPubURL = publicURL
}

func (s *Service) GetCallToken(ctx context.Context, userID, conversationID string) (CallTokenResponse, error) {
	if err := s.ensureConversation(ctx, userID, conversationID); err != nil {
		return CallTokenResponse{}, err
	}
	if s.tokens == nil {
		return CallTokenResponse{}, apperr.InvalidState
	}
	var name string
	if s.users != nil {
		if u, err := s.users.GetUserByID(ctx, userID); err == nil && u != nil {
			name = u.DisplayName
		}
	}
	if name == "" {
		name = userID
	}
	token, err := s.tokens.GenerateToken(conversationID, userID, name)
	if err != nil {
		return CallTokenResponse{}, err
	}
	return CallTokenResponse{
		Token: token,
		URL:   s.callPubURL,
		Room:  conversationID,
	}, nil
}

func (s *Service) SendCallSignal(ctx context.Context, userID, conversationID, action string, isVideo bool) error {
	action = strings.TrimSpace(strings.ToLower(action))
	switch action {
	case "invite", "accept", "decline", "end":
	default:
		return apperr.Validation
	}
	if err := s.ensureConversation(ctx, userID, conversationID); err != nil {
		return err
	}
	var name string
	if s.users != nil {
		if u, err := s.users.GetUserByID(ctx, userID); err == nil && u != nil {
			name = u.DisplayName
		}
	}
	if name == "" {
		name = userID
	}
	now := s.now().UTC()
	signal := CallSignal{
		ConversationID: conversationID,
		SenderID:       userID,
		SenderName:     name,
		Action:         action,
		IsVideo:        isVideo,
		Timestamp:      now.Format(time.RFC3339),
	}
	payload, err := json.Marshal(signal)
	if err != nil {
		return err
	}
	return s.repo.InsertChatEvent(ctx, conversationID, "call.signal", conversationID, payload, now)
}

func (s *Service) MarkAsRead(ctx context.Context, userID, conversationID, messageID string) error {
	if err := s.ensureConversation(ctx, userID, conversationID); err != nil {
		return err
	}
	now := s.now().UTC()
	if err := s.repo.MarkAsRead(ctx, userID, conversationID, messageID, now); err != nil {
		return err
	}
	signal := ReadSignal{
		ConversationID:    conversationID,
		UserID:            userID,
		LastReadMessageID: strings.TrimSpace(messageID),
		LastReadAt:        now.Format(time.RFC3339Nano),
	}
	payload, err := json.Marshal(signal)
	if err != nil {
		return err
	}
	return s.repo.InsertChatEvent(ctx, conversationID, "message.read", conversationID, payload, now)
}

func (s *Service) SendTyping(ctx context.Context, userID, conversationID string, typing bool) error {
	if err := s.ensureConversation(ctx, userID, conversationID); err != nil {
		return err
	}
	var name string
	if s.users != nil {
		if u, err := s.users.GetUserByID(ctx, userID); err == nil && u != nil {
			name = u.DisplayName
		}
	}
	if name == "" {
		name = userID
	}
	now := s.now().UTC()
	signal := TypingSignal{
		ConversationID: conversationID,
		UserID:         userID,
		UserName:       name,
		Typing:         typing,
	}
	payload, err := json.Marshal(signal)
	if err != nil {
		return err
	}
	return s.repo.InsertChatEvent(ctx, conversationID, "typing.indicator", conversationID, payload, now)
}

