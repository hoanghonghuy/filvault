package chat_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
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

func TestChat_MessagePaginationExactLimitHasNoMore(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Exact paging"})
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	for _, text := range []string{"first", "second"} {
		code, body := postAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages", token, map[string]any{"body": text})
		if code != http.StatusCreated {
			t.Fatalf("send %s status=%d body=%s", text, code, body)
		}
	}

	code, body := getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages?limit=2", token)
	if code != http.StatusOK {
		t.Fatalf("exact page status=%d body=%s", code, body)
	}
	var page struct {
		Messages []struct {
			Body string `json:"body"`
		} `json:"messages"`
		HasMore bool `json:"hasMore"`
	}
	decodeJSON(t, body, &page)
	if len(page.Messages) != 2 || page.HasMore {
		t.Fatalf("exact page should not have more: %+v body=%s", page, body)
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
	if code != http.StatusOK {
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

func TestChat_ConcurrentCompleteCreatesOneMessage(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Concurrent"})
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	code, body := postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
		"conversationId": conv.ID,
		"name":           "concurrent.jpg",
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

	type result struct {
		code int
		body string
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			code, body := postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{
				"conversationId": conv.ID,
				"body":           "one logical message",
			})
			results <- result{code: code, body: body}
		}()
	}
	wg.Wait()
	close(results)

	var messageID string
	createdCount := 0
	for result := range results {
		if result.code != http.StatusCreated && result.code != http.StatusOK {
			t.Fatalf("concurrent complete status=%d body=%s", result.code, result.body)
		}
		if result.code == http.StatusCreated {
			createdCount++
		}
		var message struct {
			ID string `json:"id"`
		}
		decodeJSON(t, result.body, &message)
		if messageID == "" {
			messageID = message.ID
		} else if message.ID != messageID {
			t.Fatalf("concurrent complete created different messages: %s and %s", messageID, message.ID)
		}
	}
	if createdCount != 1 {
		t.Fatalf("concurrent complete should create exactly one message: created=%d", createdCount)
	}

	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages?limit=50", token)
	if code != http.StatusOK || strings.Count(body, `"conversationId"`) != 1 {
		t.Fatalf("concurrent complete should leave one message: status=%d body=%s", code, body)
	}
}

func TestChat_CompleteUsesObjectStorageSize(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Object size"})
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	code, body := postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
		"conversationId": conv.ID,
		"name":           "empty.txt",
		"size":           128,
		"contentType":    "text/plain",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objs.PutObject(file.ObjectKey(userID, session.FileID), objectstore.ObjectStat{Size: 0, ContentType: "text/plain"})

	code, body = postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{
		"conversationId": conv.ID,
	})
	if code != http.StatusCreated {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	var message struct {
		Attachments []struct {
			SizeBytes int64 `json:"sizeBytes"`
		} `json:"attachments"`
	}
	decodeJSON(t, body, &message)
	if len(message.Attachments) != 1 || message.Attachments[0].SizeBytes != 0 {
		t.Fatalf("attachment size should come from object storage: %+v", message.Attachments)
	}
}

func TestChat_DirectConversationSupportsBothMembers(t *testing.T) {
	engine, memA, _ := newEngine(t)
	tokenA := registerVerified(t, engine, memA, uniqueEmail())
	tokenB := registerVerified(t, engine, memA, uniqueEmail())

	var recipient struct {
		Email string `json:"email"`
	}
	code, body := getAuth(t, engine, "/api/v1/users/me", tokenB)
	if code != http.StatusOK {
		t.Fatalf("recipient profile status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &recipient)

	code, body = postAuth(t, engine, "/api/v1/chat/direct-conversations", tokenA, map[string]any{
		"recipientEmail": recipient.Email,
	})
	if code != http.StatusOK {
		t.Fatalf("create direct conversation status=%d body=%s", code, body)
	}
	var conversation struct {
		ID   string `json:"id"`
		Peer struct {
			ID string `json:"id"`
		} `json:"peer"`
	}
	decodeJSON(t, body, &conversation)
	if conversation.ID == "" || conversation.Peer.ID == "" {
		t.Fatalf("direct conversation payload incomplete: %s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/chat/conversations", tokenB)
	if code != http.StatusOK || !strings.Contains(body, conversation.ID) {
		t.Fatalf("recipient should see direct conversation: status=%d body=%s", code, body)
	}

	code, body = postAuth(t, engine, "/api/v1/chat/conversations/"+conversation.ID+"/messages", tokenA, map[string]any{
		"clientMessageId": "01KCHATMESSAGEID0000000000001",
		"body":            "hello recipient",
	})
	if code != http.StatusCreated {
		t.Fatalf("send direct message status=%d body=%s", code, body)
	}
	if !strings.Contains(body, `"senderId"`) {
		t.Fatalf("message should expose senderId: %s", body)
	}

	code, body = postAuth(t, engine, "/api/v1/chat/conversations/"+conversation.ID+"/messages", tokenA, map[string]any{
		"clientMessageId": "01KCHATMESSAGEID0000000000001",
		"body":            "hello recipient",
	})
	if code != http.StatusOK {
		t.Fatalf("idempotent retry status=%d body=%s", code, body)
	}

	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conversation.ID+"/messages", tokenB)
	if code != http.StatusOK || strings.Count(body, "hello recipient") != 1 {
		t.Fatalf("recipient should read one message: status=%d body=%s", code, body)
	}
}

func TestChat_ConcurrentDirectConversationCreateIsCanonical(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())

	var recipient struct {
		Email string `json:"email"`
	}
	code, body := getAuth(t, engine, "/api/v1/users/me", tokenB)
	if code != http.StatusOK {
		t.Fatalf("recipient profile status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &recipient)

	type result struct {
		code int
		body string
	}
	results := make(chan result, 2)
	for _, token := range []string{tokenA, tokenA} {
		go func(token string) {
			code, body := postAuth(t, engine, "/api/v1/chat/direct-conversations", token, map[string]any{
				"recipientEmail": recipient.Email,
			})
			results <- result{code: code, body: body}
		}(token)
	}

	var conversationIDs []string
	for range 2 {
		select {
		case result := <-results:
			if result.code != http.StatusOK {
				t.Fatalf("concurrent direct conversation status=%d body=%s", result.code, result.body)
			}
			var conversation struct {
				ID string `json:"id"`
			}
			decodeJSON(t, result.body, &conversation)
			conversationIDs = append(conversationIDs, conversation.ID)
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent direct conversation creation timed out")
		}
	}
	if len(conversationIDs) != 2 || conversationIDs[0] == "" || conversationIDs[0] != conversationIDs[1] {
		t.Fatalf("direct conversation IDs are not canonical: %v", conversationIDs)
	}
}

func TestChat_DirectMessageEditRemoveAndScopedDownload(t *testing.T) {
	engine, mem, objs := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())

	var recipient struct {
		Email string `json:"email"`
	}
	code, body := getAuth(t, engine, "/api/v1/users/me", tokenB)
	if code != http.StatusOK {
		t.Fatalf("recipient profile status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &recipient)
	code, body = postAuth(t, engine, "/api/v1/chat/direct-conversations", tokenA, map[string]any{
		"recipientEmail": recipient.Email,
	})
	if code != http.StatusOK {
		t.Fatalf("create direct conversation status=%d body=%s", code, body)
	}
	var conversation struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conversation)

	code, body = postAuth(t, engine, "/api/v1/chat/conversations/"+conversation.ID+"/messages", tokenA, map[string]any{
		"clientMessageId": "01KCHATMESSAGEID0000000000002",
		"body":            "before edit",
	})
	if code != http.StatusCreated {
		t.Fatalf("send message status=%d body=%s", code, body)
	}
	var message struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &message)

	code, body = doJSON(t, engine, http.MethodPatch, "/api/v1/chat/conversations/"+conversation.ID+"/messages/"+message.ID, tokenA, map[string]any{
		"body": "after edit",
	})
	if code != http.StatusOK || !strings.Contains(body, `"editedAt"`) || !strings.Contains(body, "after edit") {
		t.Fatalf("edit message status=%d body=%s", code, body)
	}

	code, body = doJSON(t, engine, http.MethodDelete, "/api/v1/chat/conversations/"+conversation.ID+"/messages/"+message.ID, tokenB, nil)
	if code != http.StatusForbidden {
		t.Fatalf("recipient must not remove sender message: status=%d body=%s", code, body)
	}
	code, body = doJSON(t, engine, http.MethodDelete, "/api/v1/chat/conversations/"+conversation.ID+"/messages/"+message.ID, tokenA, nil)
	if code != http.StatusNoContent {
		t.Fatalf("sender remove status=%d body=%s", code, body)
	}

	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conversation.ID+"/messages", tokenB)
	if code != http.StatusOK || !strings.Contains(body, `"removedAt"`) || strings.Contains(body, "after edit") {
		t.Fatalf("removed tombstone should hide body: status=%d body=%s", code, body)
	}
	_ = objs
}

func TestChat_DirectConversationCreatesSharedAlbumForAttachment(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Album"})
	var conversation struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conversation)
	code, body := postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
		"conversationId": conversation.ID, "name": "shared.jpg", "size": 12, "contentType": "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objs.PutObject(file.ObjectKey(userID, session.FileID), objectstore.ObjectStat{Size: 12, ContentType: "image/jpeg"})
	code, body = postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{
		"conversationId": conversation.ID, "body": "album item",
	})
	if code != http.StatusCreated {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	var message struct {
		ID          string `json:"id"`
		Attachments []struct {
			ID string `json:"id"`
		} `json:"attachments"`
	}
	decodeJSON(t, body, &message)
	if len(message.Attachments) != 1 {
		t.Fatalf("attachment missing from message: %s", body)
	}
	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conversation.ID+"/messages", token)
	if code != http.StatusOK || !strings.Contains(body, message.Attachments[0].ID) {
		t.Fatalf("message attachment missing from history: status=%d body=%s", code, body)
	}
	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conversation.ID+"/attachments/"+message.Attachments[0].ID+"/download", token)
	if code != http.StatusOK || !strings.Contains(body, "downloadUrl") {
		t.Fatalf("scoped attachment download status=%d body=%s", code, body)
	}
}

func TestChat_AttachmentMessageWritesRealtimeEvent(t *testing.T) {
	engine, mem, objs := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())

	var recipient struct {
		Email string `json:"email"`
	}
	code, body := getAuth(t, engine, "/api/v1/users/me", tokenB)
	if code != http.StatusOK {
		t.Fatalf("recipient profile status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &recipient)
	code, body = postAuth(t, engine, "/api/v1/chat/direct-conversations", tokenA, map[string]any{
		"recipientEmail": recipient.Email,
	})
	if code != http.StatusOK {
		t.Fatalf("create direct conversation status=%d body=%s", code, body)
	}
	var conversation struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conversation)

	code, body = postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", tokenA, map[string]any{
		"conversationId": conversation.ID, "name": "event.jpg", "size": 12, "contentType": "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, tokenA)
	objs.PutObject(file.ObjectKey(userID, session.FileID), objectstore.ObjectStat{Size: 12, ContentType: "image/jpeg"})
	code, body = postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", tokenA, map[string]any{
		"conversationId": conversation.ID,
	})
	if code != http.StatusCreated {
		t.Fatalf("complete status=%d body=%s", code, body)
	}

	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v1/chat/events?after=0", nil)
	if err != nil {
		t.Fatalf("events request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenB)
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("events response: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("events status=%d", resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	foundEvent := false
	for scanner.Scan() {
		if scanner.Text() == "event: message.created" {
			foundEvent = true
			break
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read events: %v", err)
	}
	if !foundEvent {
		t.Fatal("attachment message should publish chat event")
	}
}

func TestChat_MediaKeepsTrashedAttachmentUnavailable(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Unavailable media"})
	var conversation struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conversation)

	code, body := postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
		"conversationId": conversation.ID, "name": "unavailable.jpg", "size": 12, "contentType": "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objs.PutObject(file.ObjectKey(userID, session.FileID), objectstore.ObjectStat{Size: 12, ContentType: "image/jpeg"})
	code, body = postAuth(t, engine, "/api/v1/chat/attachments/"+session.FileID+"/complete", token, map[string]any{
		"conversationId": conversation.ID,
	})
	if code != http.StatusCreated {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	var message struct {
		Attachments []struct {
			ID string `json:"id"`
		} `json:"attachments"`
	}
	decodeJSON(t, body, &message)
	if len(message.Attachments) != 1 {
		t.Fatalf("attachment missing from message: %s", body)
	}

	if code, body := deleteAuth(t, engine, "/api/v1/files/"+session.FileID, token); code != http.StatusNoContent {
		t.Fatalf("trash status=%d body=%s", code, body)
	}
	code, body = getAuth(t, engine, "/api/v1/chat/conversations/"+conversation.ID+"/media", token)
	if code != http.StatusOK {
		t.Fatalf("media status=%d body=%s", code, body)
	}
	var mediaResp struct {
		Media []struct {
			Availability string `json:"availability"`
		} `json:"media"`
	}
	decodeJSON(t, body, &mediaResp)
	if len(mediaResp.Media) != 1 || mediaResp.Media[0].Availability != "trashed" {
		t.Fatalf("trashed attachment should remain unavailable in shared media: %s", body)
	}

	code, _ = getAuth(t, engine, "/api/v1/chat/conversations/"+conversation.ID+"/attachments/"+message.Attachments[0].ID+"/download", token)
	if code != http.StatusNotFound {
		t.Fatalf("trashed attachment download status=%d", code)
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

func deleteAuth(t *testing.T, engine http.Handler, path, token string) (int, string) {
	return doJSON(t, engine, http.MethodDelete, path, token, nil)
}

func patchAuth(t *testing.T, engine http.Handler, path, token string, payload map[string]any) (int, string) {
	return doJSON(t, engine, http.MethodPatch, path, token, payload)
}

func TestChat_ReadReceiptAndTypingIndicator(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	emailB := uniqueEmail()
	tokenB := registerVerified(t, engine, mem, emailB)

	// User A creates direct chat with User B
	code, body := postAuth(t, engine, "/api/v1/chat/direct-conversations", tokenA, map[string]any{
		"recipientEmail": emailB,
	})
	if code != http.StatusOK {
		t.Fatalf("create direct conversation status=%d body=%s", code, body)
	}
	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conv)

	// User A sends typing signal
	code, _ = postAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/typing", tokenA, map[string]any{
		"typing": true,
	})
	if code != http.StatusOK {
		t.Fatalf("send typing status=%d", code)
	}

	// User A sends a message
	code, body = postAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/messages", tokenA, map[string]any{
		"body": "Hi there from A",
	})
	if code != http.StatusCreated {
		t.Fatalf("send message status=%d body=%s", code, body)
	}
	var msg struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &msg)

	// User B lists conversations: unreadCount should be 1
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenB)
	if code != http.StatusOK {
		t.Fatalf("list convs B status=%d body=%s", code, body)
	}
	var listB struct {
		Conversations []struct {
			ID          string `json:"id"`
			UnreadCount int    `json:"unreadCount"`
		} `json:"conversations"`
	}
	decodeJSON(t, body, &listB)
	if len(listB.Conversations) == 0 || listB.Conversations[0].UnreadCount != 1 {
		t.Fatalf("expected B unreadCount=1, got %+v", listB)
	}

	// User A lists conversations: unreadCount should be 0 (A's own message)
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenA)
	if code != http.StatusOK {
		t.Fatalf("list convs A status=%d body=%s", code, body)
	}
	var listA struct {
		Conversations []struct {
			ID          string `json:"id"`
			UnreadCount int    `json:"unreadCount"`
		} `json:"conversations"`
	}
	decodeJSON(t, body, &listA)
	if len(listA.Conversations) == 0 || listA.Conversations[0].UnreadCount != 0 {
		t.Fatalf("expected A unreadCount=0, got %+v", listA)
	}

	// User B marks as read
	code, _ = postAuth(t, engine, "/api/v1/chat/conversations/"+conv.ID+"/read", tokenB, map[string]any{
		"messageId": msg.ID,
	})
	if code != http.StatusOK {
		t.Fatalf("mark as read status=%d", code)
	}

	// User B lists conversations again: unreadCount should now be 0
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenB)
	if code != http.StatusOK {
		t.Fatalf("list convs B status=%d", code)
	}
	decodeJSON(t, body, &listB)
	if listB.Conversations[0].UnreadCount != 0 {
		t.Fatalf("expected B unreadCount=0 after reading, got %d", listB.Conversations[0].UnreadCount)
	}

	// User A lists conversations: peerLastReadMessageId should match msg.ID
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenA)
	if code != http.StatusOK {
		t.Fatalf("list convs A status=%d", code)
	}
	var listAAfter struct {
		Conversations []struct {
			ID                    string `json:"id"`
			PeerLastReadMessageID string `json:"peerLastReadMessageId"`
		} `json:"conversations"`
	}
	decodeJSON(t, body, &listAAfter)
	if listAAfter.Conversations[0].PeerLastReadMessageID != msg.ID {
		t.Fatalf("expected A to see peerLastReadMessageId=%s, got %s", msg.ID, listAAfter.Conversations[0].PeerLastReadMessageID)
	}
}

func TestChat_RealtimePresence(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())

	var recipient struct {
		Email string `json:"email"`
	}
	code, body := getAuth(t, engine, "/api/v1/users/me", tokenB)
	if code != http.StatusOK {
		t.Fatalf("recipient profile status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &recipient)

	// Create direct conversation
	code, body = postAuth(t, engine, "/api/v1/chat/direct-conversations", tokenA, map[string]any{
		"recipientEmail": recipient.Email,
	})
	if code != http.StatusOK {
		t.Fatalf("create direct conversation status=%d body=%s", code, body)
	}

	// Step 1: User B is not connected -> peerStatus must be offline
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenA)
	if code != http.StatusOK {
		t.Fatalf("list conversations status=%d body=%s", code, body)
	}
	var listResp struct {
		Conversations []struct {
			ID             string  `json:"id"`
			PeerStatus     string  `json:"peerStatus"`
			PeerLastSeenAt *string `json:"peerLastSeenAt"`
		} `json:"conversations"`
	}
	decodeJSON(t, body, &listResp)
	if len(listResp.Conversations) == 0 {
		t.Fatalf("expected conversation, got 0")
	}
	if listResp.Conversations[0].PeerStatus != "offline" {
		t.Fatalf("expected peerStatus offline, got %s", listResp.Conversations[0].PeerStatus)
	}

	// Step 2: Connect User B via SSE
	server := httptest.NewServer(engine)
	defer server.Close()

	ctxB, cancelB := context.WithCancel(context.Background())
	reqB, err := http.NewRequestWithContext(ctxB, http.MethodGet, server.URL+"/api/v1/chat/events?after=0", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	reqB.Header.Set("Authorization", "Bearer "+tokenB)
	respB, err := server.Client().Do(reqB)
	if err != nil {
		t.Fatalf("do req: %v", err)
	}
	defer respB.Body.Close()

	// Give a moment for connection handler to register
	time.Sleep(100 * time.Millisecond)

	// User A lists conversations: peerStatus should now be online
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenA)
	if code != http.StatusOK {
		t.Fatalf("list conversations status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &listResp)
	if listResp.Conversations[0].PeerStatus != "online" {
		t.Fatalf("expected peerStatus online, got %s", listResp.Conversations[0].PeerStatus)
	}

	// Step 3: Disconnect User B
	cancelB()
	_ = respB.Body.Close()

	// Give a moment for deferred HandleDisconnect to execute
	time.Sleep(150 * time.Millisecond)

	// User A lists conversations: peerStatus should now be offline with peerLastSeenAt populated
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenA)
	if code != http.StatusOK {
		t.Fatalf("list conversations status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &listResp)
	if listResp.Conversations[0].PeerStatus != "offline" {
		t.Fatalf("expected peerStatus offline after disconnect, got %s", listResp.Conversations[0].PeerStatus)
	}
	if listResp.Conversations[0].PeerLastSeenAt == nil || *listResp.Conversations[0].PeerLastSeenAt == "" {
		t.Fatalf("expected peerLastSeenAt to be set after disconnect, got nil")
	}
}

func TestChat_MutualActiveStatusPrivacy(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())

	var recipient struct {
		Email string `json:"email"`
	}
	code, body := getAuth(t, engine, "/api/v1/users/me", tokenB)
	if code != http.StatusOK {
		t.Fatalf("recipient status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &recipient)

	// User A creates direct conversation with User B
	code, _ = postAuth(t, engine, "/api/v1/chat/direct-conversations", tokenA, map[string]any{
		"recipientEmail": recipient.Email,
	})
	if code != http.StatusOK {
		t.Fatalf("direct conv status=%d", code)
	}

	server := httptest.NewServer(engine)
	defer server.Close()

	// Connect User B via SSE
	ctxB, cancelB := context.WithCancel(context.Background())
	defer cancelB()
	reqB, _ := http.NewRequestWithContext(ctxB, http.MethodGet, server.URL+"/api/v1/chat/events?after=0", nil)
	reqB.Header.Set("Authorization", "Bearer "+tokenB)
	respB, err := server.Client().Do(reqB)
	if err != nil {
		t.Fatalf("respB: %v", err)
	}
	defer respB.Body.Close()
	time.Sleep(100 * time.Millisecond)

	// Step 1: Both have active status enabled (default true) -> User A sees User B online
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenA)
	if code != http.StatusOK {
		t.Fatalf("list convs A status=%d", code)
	}
	var listResp struct {
		Conversations []struct {
			PeerStatus     string  `json:"peerStatus"`
			PeerLastSeenAt *string `json:"peerLastSeenAt"`
		} `json:"conversations"`
	}
	decodeJSON(t, body, &listResp)
	if len(listResp.Conversations) == 0 || listResp.Conversations[0].PeerStatus != "online" {
		t.Fatalf("expected B online to A, got %+v", listResp)
	}

	// Step 2: User A turns active status OFF
	code, _ = patchAuth(t, engine, "/api/v1/users/me", tokenA, map[string]any{
		"activeStatusEnabled": false,
	})
	if code != http.StatusOK {
		t.Fatalf("patch user A status=%d", code)
	}

	// Mutual privacy: User A cannot see User B's status anymore!
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenA)
	if code != http.StatusOK {
		t.Fatalf("list convs A status=%d", code)
	}
	listResp.Conversations = nil
	decodeJSON(t, body, &listResp)
	if listResp.Conversations[0].PeerStatus != "" {
		t.Fatalf("expected A to NOT see B online when A has active status off, got %s", listResp.Conversations[0].PeerStatus)
	}

	// User B also cannot see User A's status
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenB)
	if code != http.StatusOK {
		t.Fatalf("list convs B status=%d", code)
	}
	listResp.Conversations = nil
	decodeJSON(t, body, &listResp)
	if listResp.Conversations[0].PeerStatus != "" {
		t.Fatalf("expected B to NOT see A online when A has active status off, got %s", listResp.Conversations[0].PeerStatus)
	}

	// Step 3: User A turns active status back ON
	code, _ = patchAuth(t, engine, "/api/v1/users/me", tokenA, map[string]any{
		"activeStatusEnabled": true,
	})
	if code != http.StatusOK {
		t.Fatalf("patch user A back on status=%d", code)
	}

	// User A sees User B online again!
	code, body = getAuth(t, engine, "/api/v1/chat/conversations?includePreview=true", tokenA)
	if code != http.StatusOK {
		t.Fatalf("list convs A status=%d", code)
	}
	listResp.Conversations = nil
	decodeJSON(t, body, &listResp)
	if listResp.Conversations[0].PeerStatus != "online" {
		t.Fatalf("expected A to see B online again, got %s", listResp.Conversations[0].PeerStatus)
	}
}

