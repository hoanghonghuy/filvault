package chat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"filvault/internal/app"
	"filvault/internal/file"
	"filvault/internal/platform/config"
	"filvault/internal/platform/mailer"
	"filvault/internal/platform/objectstore"
	"filvault/internal/platform/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestChat_CreateConversationAndTextMessage(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Family"})
	if code != http.StatusCreated {
		t.Fatalf("create conversation status=%d body=%s", code, body)
	}
	var conv struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	decodeJSON(t, body, &conv)
	if conv.ID == "" || conv.Title != "Family" {
		t.Fatalf("conversation mismatch: %+v body=%s", conv, body)
	}

	code, body = postAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages", token, map[string]any{
		"body": "hello chat",
	})
	if code != http.StatusCreated {
		t.Fatalf("create message status=%d body=%s", code, body)
	}
	if !strings.Contains(body, "hello chat") {
		t.Fatalf("message body missing: %s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages", token)
	if code != http.StatusOK {
		t.Fatalf("list messages status=%d body=%s", code, body)
	}
	if !strings.Contains(body, "hello chat") {
		t.Fatalf("listed message missing: %s", body)
	}
}

func TestChat_MessagePaginationAndPreview(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Paging"})
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	for _, text := range []string{"msg-1", "msg-2", "msg-3"} {
		code, body := postAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages", token, map[string]any{"body": text})
		if code != http.StatusCreated {
			t.Fatalf("send %s status=%d body=%s", text, code, body)
		}
	}

	code, body := getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", token)
	if code != http.StatusOK {
		t.Fatalf("list with preview status=%d body=%s", code, body)
	}
	if !strings.Contains(body, `"preview"`) || !strings.Contains(body, "msg-3") || !strings.Contains(body, `"lastMessageAt"`) {
		t.Fatalf("conversation preview missing: %s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages?limit=2", token)
	if code != http.StatusOK {
		t.Fatalf("latest page status=%d body=%s", code, body)
	}
	var latest struct {
		Messages []struct {
			Body string `json:"body"`
		} `json:"messages"`
		HasMore    bool   `json:"hasMore"`
		NextBefore string `json:"nextBefore"`
	}
	decodeJSON(t, body, &latest)
	if len(latest.Messages) != 2 || latest.Messages[0].Body != "msg-2" || latest.Messages[1].Body != "msg-3" {
		t.Fatalf("latest window wrong: %+v", latest.Messages)
	}
	if !latest.HasMore || latest.NextBefore == "" {
		t.Fatalf("expected hasMore + cursor: %+v %+v", latest.HasMore, latest.NextBefore)
	}

	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages?limit=2&before="+latest.NextBefore, token)
	if code != http.StatusOK {
		t.Fatalf("older page status=%d body=%s", code, body)
	}
	var older struct {
		Messages []struct {
			Body string `json:"body"`
		} `json:"messages"`
		HasMore bool `json:"hasMore"`
	}
	decodeJSON(t, body, &older)
	if len(older.Messages) != 1 || older.Messages[0].Body != "msg-1" {
		t.Fatalf("older page wrong: %+v", older.Messages)
	}
	if older.HasMore {
		t.Fatalf("older page should not have more")
	}

	if code, body := getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages?limit=0", token); code != http.StatusBadRequest {
		t.Fatalf("limit=0 status=%d body=%s", code, body)
	}
	if code, body := getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages?before=not-a-ulid", token); code != http.StatusBadRequest {
		t.Fatalf("bad cursor status=%d body=%s", code, body)
	}
}

func TestChat_ImageAttachmentAppearsInPhotos(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Images"})
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	code, body := postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
		"conversationId": conv.ID,
		"name":           "chat-sunset.jpg",
		"size":           128,
		"contentType":    "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objs.PutObject(file.ObjectKey(userID, session.FileID), objectstore.ObjectStat{Size: 128, ContentType: "image/jpeg"})

	code, body = postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{
		"conversationId": conv.ID,
		"body":           "sent from chat",
	})
	if code != http.StatusCreated {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	if !strings.Contains(body, "chat-sunset.jpg") {
		t.Fatalf("attachment name missing: %s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/timeline", token)
	if code != http.StatusOK {
		t.Fatalf("timeline status=%d body=%s", code, body)
	}
	if !strings.Contains(body, "chat-sunset.jpg") {
		t.Fatalf("chat image must appear in Photos timeline: %s", body)
	}
}

func TestChat_SearchMessagesAndMediaGallery(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Search"})
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	postAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages", token, map[string]any{"body": "alpha plain"})
	postAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages", token, map[string]any{"body": "family sunset memory"})

	code, body := getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages/search?q=sunset", token)
	if code != http.StatusOK {
		t.Fatalf("search status=%d body=%s", code, body)
	}
	if !strings.Contains(body, "family sunset memory") || strings.Contains(body, "alpha plain") {
		t.Fatalf("search result mismatch: %s", body)
	}

	code, body = postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
		"conversationId": conv.ID,
		"name":           "gallery-sunset.jpg",
		"size":           256,
		"contentType":    "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objs.PutObject(file.ObjectKey(userID, session.FileID), objectstore.ObjectStat{Size: 256, ContentType: "image/jpeg"})
	postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{"conversationId": conv.ID})

	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/media", token)
	if code != http.StatusOK {
		t.Fatalf("media status=%d body=%s", code, body)
	}
	if !strings.Contains(body, "gallery-sunset.jpg") || !strings.Contains(body, "image/jpeg") {
		t.Fatalf("media gallery mismatch: %s", body)
	}
}

func TestChat_MediaThumbnails(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Thumbs"})
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	for _, name := range []string{"pic.jpg", "doc.pdf"} {
		code, body := postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
			"conversationId": conv.ID,
			"name":           name,
			"size":           64,
			"contentType":    map[string]string{"pic.jpg": "image/jpeg", "doc.pdf": "application/pdf"}[name],
		})
		if code != http.StatusCreated {
			t.Fatalf("session %s status=%d", name, code)
		}
		var session struct {
			FileID string `json:"fileId"`
		}
		decodeJSON(t, body, &session)
		userID := decodeUserID(t, engine, token)
		objs.PutObject(file.ObjectKey(userID, session.FileID), objectstore.ObjectStat{Size: 64, ContentType: map[string]string{"pic.jpg": "image/jpeg", "doc.pdf": "application/pdf"}[name]})
		code, _ = postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{"conversationId": conv.ID})
		if code != http.StatusCreated {
			t.Fatalf("complete %s status=%d", name, code)
		}
	}

	code, body := getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/media", token)
	if code != http.StatusOK {
		t.Fatalf("media status=%d body=%s", code, body)
	}
	var mediaResp struct {
		Media []struct {
			Name         string `json:"name"`
			MimeType     string `json:"mimeType"`
			ThumbnailURL string `json:"thumbnailUrl"`
		} `json:"media"`
	}
	decodeJSON(t, body, &mediaResp)
	if len(mediaResp.Media) != 1 || mediaResp.Media[0].Name != "pic.jpg" {
		t.Fatalf("gallery should contain only the image: %s", body)
	}
	if mediaResp.Media[0].ThumbnailURL == "" {
		t.Fatalf("image should have thumbnailUrl: %s", body)
	}
}

func TestChat_MessageAttachmentThumbnails(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Inline"})
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	code, body := postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
		"conversationId": conv.ID,
		"name":           "inline.jpg",
		"size":           64,
		"contentType":    "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objs.PutObject(file.ObjectKey(userID, session.FileID), objectstore.ObjectStat{Size: 64, ContentType: "image/jpeg"})
	code, _ = postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{"conversationId": conv.ID})
	if code != http.StatusCreated {
		t.Fatalf("complete status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages?limit=50", token)
	if code != http.StatusOK {
		t.Fatalf("messages status=%d body=%s", code, body)
	}
	var page struct {
		Messages []struct {
			Attachments []struct {
				Name         string `json:"name"`
				MimeType     string `json:"mimeType"`
				ThumbnailURL string `json:"thumbnailUrl"`
			} `json:"attachments"`
		} `json:"messages"`
	}
	decodeJSON(t, body, &page)
	if len(page.Messages) != 1 || len(page.Messages[0].Attachments) != 1 {
		t.Fatalf("expected one message with one attachment: %s", body)
	}
	a := page.Messages[0].Attachments[0]
	if a.Name != "inline.jpg" || a.MimeType != "image/jpeg" {
		t.Fatalf("wrong attachment: %+v", a)
	}
	if a.ThumbnailURL == "" {
		t.Fatalf("image attachment should have thumbnailUrl in message payload: %s", body)
	}
}

func TestChat_CompleteAttachmentIsIdempotent(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Retry"})
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	code, body := postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
		"conversationId": conv.ID,
		"name":           "retry.jpg",
		"size":           128,
		"contentType":    "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objs.PutObject(file.ObjectKey(userID, session.FileID), objectstore.ObjectStat{Size: 128, ContentType: "image/jpeg"})

	code, body = postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{
		"conversationId": conv.ID,
		"body":           "retry me",
	})
	if code != http.StatusCreated {
		t.Fatalf("first complete status=%d body=%s", code, body)
	}
	var first struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &first)

	code, body = postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{
		"conversationId": conv.ID,
		"body":           "retry me",
	})
	if code != http.StatusCreated {
		t.Fatalf("retry complete status=%d body=%s", code, body)
	}
	var second struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &second)
	if second.ID != first.ID {
		t.Fatalf("retry created a different message: first=%s second=%s", first.ID, second.ID)
	}

	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages?limit=50", token)
	if code != http.StatusOK || strings.Count(body, `"conversationId"`) != 1 {
		t.Fatalf("retry should leave one message: status=%d body=%s", code, body)
	}
}

func newEngine(t *testing.T) (*gin.Engine, *mailer.Memory, *objectstore.Memory) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	url := os.Getenv("FILVAULT_DATABASE_URL")
	if url == "" {
		t.Fatal("FILVAULT_DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	mem := mailer.NewMemory()
	objs := objectstore.NewMemory()
	engine := app.NewWithDeps(config.Config{
		InviteCode:                "secret-invite",
		JWTSecret:                 "test-jwt-secret-not-for-prod",
		DefaultTrashAutoDelete:    false,
		DefaultTrashRetentionDays: config.DefaultTrashRetentionDays,
		MetadataStore:             "postgres",
	}, pool, mem, objs)
	return engine, mem, objs
}

func decodeUserID(t *testing.T, engine http.Handler, token string) string {
	t.Helper()
	code, body := getAuth(t, engine, "/api/v1/users/me", token)
	if code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", code, body)
	}
	var me struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &me)
	return me.ID
}

func registerVerified(t *testing.T, engine http.Handler, mem *mailer.Memory, email string) string {
	t.Helper()
	code, _ := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       email,
		"password":    "password1",
		"displayName": "User",
		"inviteCode":  "secret-invite",
	})
	if code != http.StatusCreated {
		t.Fatalf("register status=%d", code)
	}
	code, _ = postJSON(t, engine, "/api/v1/auth/resend-verification", map[string]any{"email": email})
	if code != http.StatusNoContent {
		t.Fatalf("resend status=%d", code)
	}
	code, _ = postJSON(t, engine, "/api/v1/auth/verify-email", map[string]any{
		"email": email,
		"code":  mem.LastCode(email),
	})
	if code != http.StatusNoContent {
		t.Fatalf("verify status=%d", code)
	}
	code, body := postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "password1",
	})
	if code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", code, body)
	}
	var login struct {
		AccessToken string `json:"accessToken"`
	}
	decodeJSON(t, body, &login)
	return login.AccessToken
}

func uniqueEmail() string {
	return fmt.Sprintf("chat-%d@example.com", time.Now().UnixNano())
}

func decodeJSON(t *testing.T, body string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), v); err != nil {
		t.Fatalf("json: %v body=%s", err, body)
	}
}

func doJSON(t *testing.T, engine http.Handler, method, path, token string, payload map[string]any) (int, string) {
	t.Helper()
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func postJSON(t *testing.T, engine http.Handler, path string, payload map[string]any) (int, string) {
	return doJSON(t, engine, http.MethodPost, path, "", payload)
}

func postAuth(t *testing.T, engine http.Handler, path, token string, payload map[string]any) (int, string) {
	return doJSON(t, engine, http.MethodPost, path, token, payload)
}

func getAuth(t *testing.T, engine http.Handler, path, token string) (int, string) {
	return doJSON(t, engine, http.MethodGet, path, token, nil)
}
