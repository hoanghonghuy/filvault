package share_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestShares_CreateListRevoke(t *testing.T) {
	engine, mem, _ := newEngine(t)
	emailA := uniqueEmail("a")
	tokenA := registerVerified(t, engine, mem, emailA)
	emailB := uniqueEmail("b")
	registerVerified(t, engine, mem, emailB)

	// Create a folder owned by A.
	_, body := postAuth(t, engine, "/api/v1/folders", tokenA, map[string]any{"name": "Shared"})
	var folder folderResp
	decodeJSON(t, body, &folder)

	// Share the folder with B.
	code, body := postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   folder.ID,
		"email":        emailB,
	})
	if code != http.StatusCreated {
		t.Fatalf("share status=%d body=%s", code, body)
	}
	var created shareResp
	decodeJSON(t, body, &created)
	if created.ResourceType != "folder" || created.ResourceID != folder.ID {
		t.Fatalf("share=%+v", created)
	}

	// Duplicate share -> conflict.
	code, body = postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   folder.ID,
		"email":        emailB,
	})
	assertAPIError(t, code, body, http.StatusConflict, "CONFLICT")

	// Self-share -> validation.
	code, body = postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   folder.ID,
		"email":        emailA,
	})
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	// Outgoing list for A.
	code, body = getAuth(t, engine, "/api/v1/shares", tokenA)
	if code != http.StatusOK {
		t.Fatalf("outgoing status=%d body=%s", code, body)
	}
	var outgoing struct {
		Shares []outgoingResp `json:"shares"`
	}
	decodeJSON(t, body, &outgoing)
	if len(outgoing.Shares) != 1 || outgoing.Shares[0].ResourceName != "Shared" {
		t.Fatalf("outgoing=%s", body)
	}

	// Revoke.
	code, _ = deleteAuth(t, engine, "/api/v1/shares/"+created.ID, tokenA)
	if code != http.StatusNoContent {
		t.Fatalf("revoke status=%d", code)
	}

	// Outgoing now empty.
	code, body = getAuth(t, engine, "/api/v1/shares", tokenA)
	decodeJSON(t, body, &outgoing)
	if len(outgoing.Shares) != 0 {
		t.Fatalf("expected empty outgoing, got %s", body)
	}
}

func TestShares_IncomingAndBrowse(t *testing.T) {
	engine, mem, objs := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail("a"))
	emailB := uniqueEmail("b")
	tokenB := registerVerified(t, engine, mem, emailB)

	// A creates a folder with a child file.
	_, body := postAuth(t, engine, "/api/v1/folders", tokenA, map[string]any{"name": "Docs"})
	var folder folderResp
	decodeJSON(t, body, &folder)

	// Upload a file into the folder.
	fileID := uploadFile(t, engine, objs, tokenA, folder.ID)

	// Share folder with B.
	code, body := postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   folder.ID,
		"email":        emailB,
	})
	if code != http.StatusCreated {
		t.Fatalf("share status=%d body=%s", code, body)
	}

	// B sees incoming.
	code, body = getAuth(t, engine, "/api/v1/shares/with-me", tokenB)
	if code != http.StatusOK {
		t.Fatalf("incoming status=%d body=%s", code, body)
	}
	var incoming struct {
		Shares []incomingResp `json:"shares"`
	}
	decodeJSON(t, body, &incoming)
	if len(incoming.Shares) != 1 || incoming.Shares[0].ResourceName != "Docs" {
		t.Fatalf("incoming=%s", body)
	}

	// B browses the shared folder.
	code, body = getAuth(t, engine, "/api/v1/shared/folders/"+folder.ID, tokenB)
	if code != http.StatusOK {
		t.Fatalf("browse status=%d body=%s", code, body)
	}
	var browse struct {
		Folder *sharedFolderResp `json:"folder"`
		Files  []sharedFileResp  `json:"files"`
	}
	decodeJSON(t, body, &browse)
	if browse.Folder == nil || browse.Folder.Name != "Docs" {
		t.Fatalf("browse folder=%s", body)
	}
	if len(browse.Files) != 1 || browse.Files[0].ID != fileID {
		t.Fatalf("browse files=%s", body)
	}

	// B downloads the shared file.
	code, body = getAuth(t, engine, "/api/v1/shared/files/"+fileID+"/download", tokenB)
	if code != http.StatusOK {
		t.Fatalf("download status=%d body=%s", code, body)
	}
	var dl struct {
		DownloadURL string `json:"downloadUrl"`
	}
	decodeJSON(t, body, &dl)
	if dl.DownloadURL == "" {
		t.Fatalf("download url empty: %s", body)
	}
}

func TestShares_UnknownEmailNotFound(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail("a"))

	_, body := postAuth(t, engine, "/api/v1/folders", tokenA, map[string]any{"name": "X"})
	var folder folderResp
	decodeJSON(t, body, &folder)

	code, body := postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   folder.ID,
		"email":        "nobody@example.com",
	})
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
}

func TestShares_RevokeByNonOwnerNotFound(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail("a"))
	emailB := uniqueEmail("b")
	tokenB := registerVerified(t, engine, mem, emailB)

	_, body := postAuth(t, engine, "/api/v1/folders", tokenA, map[string]any{"name": "Y"})
	var folder folderResp
	decodeJSON(t, body, &folder)

	_, body = postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   folder.ID,
		"email":        emailB,
	})
	var created shareResp
	decodeJSON(t, body, &created)

	// B tries to revoke A's share -> not found.
	code, body := deleteAuth(t, engine, "/api/v1/shares/"+created.ID, tokenB)
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
}

type shareResp struct {
	ID           string `json:"id"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
}

type outgoingResp struct {
	ID           string  `json:"id"`
	ResourceType string  `json:"resourceType"`
	ResourceID   string  `json:"resourceId"`
	ResourceName string  `json:"resourceName"`
	Recipient    userRef `json:"recipient"`
}

type incomingResp struct {
	ID           string  `json:"id"`
	ResourceType string  `json:"resourceType"`
	ResourceID   string  `json:"resourceId"`
	ResourceName string  `json:"resourceName"`
	Owner        userRef `json:"owner"`
}

type userRef struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type folderResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type sharedFolderResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type sharedFileResp struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}

func uploadFile(t *testing.T, engine http.Handler, objs *objectstore.Memory, token, folderID string) string {
	t.Helper()
	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        "note.txt",
		"size":        4,
		"contentType": "text/plain",
		"folderId":    folderID,
	})
	if code != http.StatusCreated {
		t.Fatalf("upload-session status=%d body=%s", code, body)
	}
	var s struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &s)
	userID := decodeUserID(t, engine, token)
	objectKey := file.ObjectKey(userID, s.FileID)
	objs.PutObject(objectKey, objectstore.ObjectStat{Size: 4, ContentType: "text/plain"})
	code, body = postAuth(t, engine, "/api/v1/files/"+s.FileID+"/complete", token, nil)
	if code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	return s.FileID
}

func emailAOf(t *testing.T, token string) string {
	t.Helper()
	// We don't have the email directly; derive from a fresh register is not
	// possible. Instead, self-share is tested via a dedicated helper that
	// registers a known email. This is a placeholder to satisfy the test.
	return ""
}

func registerOnly(t *testing.T, engine http.Handler, email string) string {
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
	var s struct {
		AccessToken string `json:"accessToken"`
	}
	decodeJSON(t, body, &s)
	return s.AccessToken
}

func registerVerified(t *testing.T, engine http.Handler, mem *mailer.Memory, email string) string {
	t.Helper()
	registerOnly(t, engine, email)
	code, _ := postJSON(t, engine, "/api/v1/auth/resend-verification", map[string]any{"email": email})
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
	var s struct {
		AccessToken string `json:"accessToken"`
	}
	decodeJSON(t, body, &s)
	return s.AccessToken
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
	var u struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &u)
	return u.ID
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("share-%s-%d@example.com", prefix, time.Now().UnixNano())
}

func decodeJSON(t *testing.T, body string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), v); err != nil {
		t.Fatalf("json: %v body=%s", err, body)
	}
}

func assertAPIError(t *testing.T, status int, body string, wantStatus int, wantCode string) {
	t.Helper()
	if status != wantStatus {
		t.Fatalf("status=%d want %d body=%s", status, wantStatus, body)
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
		t.Fatalf("code=%q want %q body=%s", payload.Error.Code, wantCode, body)
	}
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
