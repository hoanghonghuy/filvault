package sharelink_test

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

// The public surface must answer 404 with one shared error code for every
// failure reason so callers cannot probe link state.
func TestShareLink_CreateDownloadRevoke(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReady(t, engine, objs, token, "share-me.jpg")

	code, body := postAuth(t, engine, "/api/v1/files/"+fileID+"/share", token, map[string]any{
		"expiresIn": "24h",
	})
	if code != http.StatusCreated {
		t.Fatalf("create share status=%d body=%s", code, body)
	}
	var created struct {
		ID        string  `json:"id"`
		FileID    string  `json:"fileId"`
		URL       string  `json:"url"`
		Token     string  `json:"token"`
		ExpiresAt *string `json:"expiresAt"`
	}
	decodeJSON(t, body, &created)
	if created.ID == "" || created.FileID != fileID {
		t.Fatalf("create payload wrong: %s", body)
	}
	if !strings.Contains(created.URL, "/s/") || strings.Contains(created.URL, created.ID) {
		t.Fatalf("url must be /s/<token> without internal id: %s", created.URL)
	}
	if created.Token == "" {
		t.Fatalf("create must return the plaintext token once: %s", body)
	}
	if created.ExpiresAt == nil {
		t.Fatalf("24h link must carry expiresAt: %s", body)
	}

	pubCode, pubBody := getJSON(t, engine, "/api/v1/public/shares/"+created.Token)
	if pubCode != http.StatusOK {
		t.Fatalf("public meta status=%d body=%s", pubCode, pubBody)
	}
	for _, leak := range []string{fileID, "ownerId", `"id"`} {
		if strings.Contains(pubBody, leak) {
			t.Fatalf("public metadata leaks %q: %s", leak, pubBody)
		}
	}
	if !strings.Contains(pubBody, "share-me.jpg") || !strings.Contains(pubBody, "image/jpeg") {
		t.Fatalf("public metadata incomplete: %s", pubBody)
	}

	dlCode, dlBody := getJSON(t, engine, "/api/v1/public/shares/"+created.Token+"/download")
	if dlCode != http.StatusOK || !strings.Contains(dlBody, "downloadUrl") {
		t.Fatalf("public download status=%d body=%s", dlCode, dlBody)
	}

	// Re-download of a fresh presign is allowed; revoke then blocks everything.
	code, _ = deleteAuth(t, engine, "/api/v1/files/"+fileID+"/share", token)
	if code != http.StatusNoContent {
		t.Fatalf("revoke status=%d", code)
	}
	assertPublicNotFound(t, engine, "/api/v1/public/shares/"+created.Token)
	assertPublicNotFound(t, engine, "/api/v1/public/shares/"+created.Token+"/download")
}

func TestShareLink_RecreateRevokesOldLink(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReady(t, engine, objs, token, "one-link.jpg")

	code, body := postAuth(t, engine, "/api/v1/files/"+fileID+"/share", token, nil)
	if code != http.StatusCreated {
		t.Fatalf("create first status=%d body=%s", code, body)
	}
	var first struct {
		Token string `json:"token"`
	}
	decodeJSON(t, body, &first)

	code, body = postAuth(t, engine, "/api/v1/files/"+fileID+"/share", token, nil)
	if code != http.StatusCreated {
		t.Fatalf("create second status=%d body=%s", code, body)
	}
	var second struct {
		Token string `json:"token"`
	}
	decodeJSON(t, body, &second)

	assertPublicNotFound(t, engine, "/api/v1/public/shares/"+first.Token)

	pubCode, _ := getJSON(t, engine, "/api/v1/public/shares/"+second.Token)
	if pubCode != http.StatusOK {
		t.Fatalf("second link must work: status=%d", pubCode)
	}
}

func TestShareLink_ListOwnerLinks(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileA := uploadReady(t, engine, objs, token, "a.jpg")
	fileB := uploadReady(t, engine, objs, token, "b.jpg")

	for _, id := range []string{fileA, fileB} {
		code, body := postAuth(t, engine, "/api/v1/files/"+id+"/share", token, nil)
		if code != http.StatusCreated {
			t.Fatalf("create %s status=%d body=%s", id, code, body)
		}
	}

	code, body := getAuth(t, engine, "/api/v1/share-links", token)
	if code != http.StatusOK {
		t.Fatalf("list links status=%d body=%s", code, body)
	}
	if !strings.Contains(body, fileA) || !strings.Contains(body, fileB) || !strings.Contains(body, "a.jpg") {
		t.Fatalf("list missing entries: %s", body)
	}
}

func TestShareLink_TrashAndPurgeKillLink(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	trashedID := uploadReady(t, engine, objs, token, "trashed.jpg")
	purgedID := uploadReady(t, engine, objs, token, "purged.jpg")

	links := make(map[string]string, 2)
	for _, id := range []string{trashedID, purgedID} {
		code, body := postAuth(t, engine, "/api/v1/files/"+id+"/share", token, nil)
		if code != http.StatusCreated {
			t.Fatalf("create %s status=%d body=%s", id, code, body)
		}
		var out struct {
			Token string `json:"token"`
		}
		decodeJSON(t, body, &out)
		links[id] = out.Token
	}

	code, _ := deleteAuth(t, engine, "/api/v1/files/"+trashedID, token)
	if code != http.StatusNoContent {
		t.Fatalf("soft delete status=%d", code)
	}
	assertPublicNotFound(t, engine, "/api/v1/public/shares/"+links[trashedID])

	// Creating a new link for a trashed file is rejected.
	code, body := postAuth(t, engine, "/api/v1/files/"+trashedID+"/share", token, nil)
	if code == http.StatusCreated {
		t.Fatalf("share trashed file must fail: %s", body)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/files/"+purgedID, token)
	if code != http.StatusNoContent {
		t.Fatalf("soft delete purge target status=%d", code)
	}
	code, _ = deleteAuth(t, engine, "/api/v1/trash/files/"+purgedID, token)
	if code != http.StatusNoContent {
		t.Fatalf("purge status=%d", code)
	}
	assertPublicNotFound(t, engine, "/api/v1/public/shares/"+links[purgedID])
}

func TestShareLink_UnknownOrForeignFileRejected(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	other := registerVerified(t, engine, mem, uniqueEmail())
	foreignID := uploadReady(t, engine, objs, other, "theirs.jpg")

	code, _ := postAuth(t, engine, "/api/v1/files/01AAAAAAAAAAAAAAAAAAAAAAAA/share", token, nil)
	if code != http.StatusNotFound {
		t.Fatalf("unknown file status=%d want=404", code)
	}
	code, _ = postAuth(t, engine, "/api/v1/files/"+foreignID+"/share", token, nil)
	if code != http.StatusNotFound {
		t.Fatalf("foreign file status=%d want=404", code)
	}
	code, body := postAuth(t, engine, "/api/v1/files/bad-ulid/share", token, nil)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	// Public probe with garbage token uses the shared 404.
	assertPublicNotFound(t, engine, "/api/v1/public/shares/not-a-real-token")
}

func assertPublicNotFound(t *testing.T, engine http.Handler, path string) {
	t.Helper()
	code, body := getJSON(t, engine, path)
	if code != http.StatusNotFound {
		t.Fatalf("%s status=%d want=404 body=%s", path, code, body)
	}
	if !strings.Contains(body, "NOT_FOUND") {
		t.Fatalf("%s must use shared NOT_FOUND code: %s", path, body)
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
		DefaultImageThumbnails:    true,
		DefaultVideoThumbnails:    true,
		DefaultTrashAutoDelete:    false,
		DefaultTrashRetentionDays: config.DefaultTrashRetentionDays,
		MetadataStore:             "postgres",
	}, pool, mem, objs)
	return engine, mem, objs
}

func uploadReady(t *testing.T, engine http.Handler, objs *objectstore.Memory, token, name string) string {
	t.Helper()
	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        name,
		"size":        128,
		"contentType": "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objectKey := file.ObjectKey(userID, session.FileID)
	objs.PutObject(objectKey, objectstore.ObjectStat{Size: 128, ContentType: "image/jpeg"})
	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", token, nil)
	if code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	return session.FileID
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
	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       email,
		"password":    "password1",
		"displayName": "User",
		"inviteCode":  "secret-invite",
	})
	if code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", code, body)
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
	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "password1",
	})
	if code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", code, body)
	}
	var s struct {
		AccessToken string `json:"accessToken"`
	}
	decodeJSON(t, body, &s)
	return s.AccessToken
}

func uniqueEmail() string {
	return fmt.Sprintf("sharelink-%d@example.com", time.Now().UnixNano())
}

func decodeJSON(t *testing.T, body string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		t.Fatalf("json: %v body=%s", err, body)
	}
}

func assertAPIError(t *testing.T, status int, body string, wantStatus int, wantCode string) {
	t.Helper()
	if status != wantStatus {
		t.Fatalf("status=%d want=%d body=%s", status, wantStatus, body)
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("error json: %v body=%s", err, body)
	}
	if payload.Error.Code != wantCode {
		t.Fatalf("code=%q want=%q body=%s", payload.Error.Code, wantCode, body)
	}
}

func doJSON(t *testing.T, engine http.Handler, method, path, token string, payload map[string]any) (int, string) {
	t.Helper()
	var buf bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&buf).Encode(payload); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func getJSON(t *testing.T, engine http.Handler, path string) (int, string) {
	return doJSON(t, engine, http.MethodGet, path, "", nil)
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
