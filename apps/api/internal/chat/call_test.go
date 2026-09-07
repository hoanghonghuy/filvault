package chat

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/file"
	"filvault/internal/platform/objectstore"
	"filvault/internal/user"
)

type mockChatRepo struct {
	getConvFn       func(ctx context.Context, ownerID, id string) (*Conversation, error)
	insertEventFn   func(ctx context.Context, conversationID, eventType, aggregateID string, payload []byte, now time.Time) error
	insertedPayload []byte
	insertedType    string
}

func (m *mockChatRepo) CreateConversation(ctx context.Context, c Conversation) error { return nil }
func (m *mockChatRepo) CreateOrGetDirectConversation(ctx context.Context, ownerID, recipientID string) (Conversation, error) {
	return Conversation{}, nil
}
func (m *mockChatRepo) ListConversations(ctx context.Context, ownerID string) ([]Conversation, error) {
	return nil, nil
}
func (m *mockChatRepo) GetConversation(ctx context.Context, ownerID, id string) (*Conversation, error) {
	if m.getConvFn != nil {
		return m.getConvFn(ctx, ownerID, id)
	}
	return &Conversation{ID: id, OwnerID: ownerID}, nil
}
func (m *mockChatRepo) CreateMessage(ctx context.Context, msg Message) (Message, error) {
	return msg, nil
}
func (m *mockChatRepo) UpdateMessage(ctx context.Context, userID, conversationID, messageID, body string, now time.Time) (Message, error) {
	return Message{}, nil
}
func (m *mockChatRepo) RemoveMessage(ctx context.Context, userID, conversationID, messageID string, now time.Time) error {
	return nil
}
func (m *mockChatRepo) CreateMessageWithAttachments(ctx context.Context, msg Message, fileIDs []string) (Message, error) {
	return msg, nil
}
func (m *mockChatRepo) ListMessages(ctx context.Context, ownerID, conversationID string, limit int) ([]Message, error) {
	return nil, nil
}
func (m *mockChatRepo) ListMessagesBefore(ctx context.Context, ownerID, conversationID, before string, limit int) ([]Message, error) {
	return nil, nil
}
func (m *mockChatRepo) LastMessageForConversation(ctx context.Context, ownerID, conversationID string) (*Message, error) {
	return nil, nil
}
func (m *mockChatRepo) SearchMessages(ctx context.Context, ownerID, conversationID, query string, limit int) ([]Message, error) {
	return nil, nil
}
func (m *mockChatRepo) CreateAttachment(ctx context.Context, a Attachment) error { return nil }
func (m *mockChatRepo) ListMedia(ctx context.Context, ownerID, conversationID string, limit int) ([]Attachment, error) {
	return nil, nil
}
func (m *mockChatRepo) GetAttachmentObjectKey(ctx context.Context, userID, conversationID, attachmentID string) (string, error) {
	return "", nil
}
func (m *mockChatRepo) ListEventsAfter(ctx context.Context, userID string, after int64, limit int) ([]Event, error) {
	return nil, nil
}
func (m *mockChatRepo) CompleteAttachment(ctx context.Context, ownerID, conversationID string, f file.File, stat objectstore.ObjectStat, body string, now time.Time) (Message, error) {
	return Message{}, nil
}
func (m *mockChatRepo) InsertChatEvent(ctx context.Context, conversationID, eventType, aggregateID string, payload []byte, now time.Time) error {
	m.insertedType = eventType
	m.insertedPayload = payload
	if m.insertEventFn != nil {
		return m.insertEventFn(ctx, conversationID, eventType, aggregateID, payload, now)
	}
	return nil
}
func (m *mockChatRepo) MarkAsRead(ctx context.Context, userID, conversationID, messageID string, now time.Time) error {
	return nil
}
func (m *mockChatRepo) ListConversationReadStates(ctx context.Context, userID string) (map[string]ReadState, error) {
	return nil, nil
}


type mockUserDir struct {
	users map[string]*user.User
}

func (m *mockUserDir) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, nil
}
func (m *mockUserDir) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	if m.users != nil {
		if u, ok := m.users[id]; ok {
			return u, nil
		}
	}
	return nil, nil
}

type mockTokenGen struct {
	token string
	err   error
}

func (m *mockTokenGen) GenerateToken(room, identity, name string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func TestService_GetCallToken_UnauthorizedIfNotMember(t *testing.T) {
	repo := &mockChatRepo{
		getConvFn: func(ctx context.Context, ownerID, id string) (*Conversation, error) {
			return nil, nil // not found
		},
	}
	svc := NewService(repo, nil, nil, nil, nil)
	svc.SetCallConfig(&mockTokenGen{token: "mock-jwt"}, "ws://localhost:7880")

	_, err := svc.GetCallToken(context.Background(), "user-1", "conv-999")
	if err != apperr.NotFound {
		t.Fatalf("expected NotFound error, got %v", err)
	}
}

func TestService_GetCallToken_Success(t *testing.T) {
	repo := &mockChatRepo{
		getConvFn: func(ctx context.Context, ownerID, id string) (*Conversation, error) {
			return &Conversation{ID: id, OwnerID: ownerID}, nil
		},
	}
	userDir := &mockUserDir{
		users: map[string]*user.User{
			"user-1": {ID: "user-1", DisplayName: "Alice"},
		},
	}
	svc := NewService(repo, userDir, nil, nil, nil)
	svc.SetCallConfig(&mockTokenGen{token: "mock-jwt-token"}, "ws://livekit.example.com:7880")

	res, err := svc.GetCallToken(context.Background(), "user-1", "conv-100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token != "mock-jwt-token" {
		t.Errorf("expected token 'mock-jwt-token', got %q", res.Token)
	}
	if res.URL != "ws://livekit.example.com:7880" {
		t.Errorf("expected URL 'ws://livekit.example.com:7880', got %q", res.URL)
	}
	if res.Room != "conv-100" {
		t.Errorf("expected Room 'conv-100', got %q", res.Room)
	}
}

func TestService_SendCallSignal_Validation(t *testing.T) {
	repo := &mockChatRepo{}
	svc := NewService(repo, nil, nil, nil, nil)

	// Invalid action
	err := svc.SendCallSignal(context.Background(), "user-1", "conv-1", "unknown_action", false)
	if err != apperr.Validation {
		t.Fatalf("expected Validation error, got %v", err)
	}
}

func TestService_SendCallSignal_Success(t *testing.T) {
	repo := &mockChatRepo{}
	userDir := &mockUserDir{
		users: map[string]*user.User{
			"user-1": {ID: "user-1", DisplayName: "Bob"},
		},
	}
	svc := NewService(repo, userDir, nil, nil, nil)

	err := svc.SendCallSignal(context.Background(), "user-1", "conv-1", "invite", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.insertedType != "call.signal" {
		t.Errorf("expected insertedType 'call.signal', got %q", repo.insertedType)
	}

	var signal CallSignal
	if err := json.Unmarshal(repo.insertedPayload, &signal); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if signal.Action != "invite" {
		t.Errorf("expected action 'invite', got %q", signal.Action)
	}
	if signal.SenderName != "Bob" {
		t.Errorf("expected senderName 'Bob', got %q", signal.SenderName)
	}
	if !signal.IsVideo {
		t.Errorf("expected isVideo to be true")
	}
	if signal.ConversationID != "conv-1" {
		t.Errorf("expected conversationId 'conv-1', got %q", signal.ConversationID)
	}
}
