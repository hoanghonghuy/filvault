package file_test

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

func TestUploadSession_RejectsInvalidType(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        "evil.exe",
		"size":        100,
		"contentType": "application/octet-stream",
	})
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func TestUploadSession_RejectsTooLarge(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        "big.jpg",
		"size":        config.MaxFileSizeBytes + 1,
		"contentType": "image/jpeg",
	})
	assertAPIError(t, code, body, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE")
}

func TestUploadComplete_BrowserDownload(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID, _ := uploadReady(t, engine, objs, token, "photo.jpg", nil)

	code, body := getAuth(t, engine, "/api/v1/browser", token)
	if code != http.StatusOK {
		t.Fatalf("browser status=%d body=%s", code, body)
	}
	if !strings.Contains(body, "photo.jpg") {
		t.Fatalf("browser missing file: %s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/files/"+fileID+"/download", token)
	if code != http.StatusOK {
		t.Fatalf("download status=%d body=%s", code, body)
	}
	var dl struct {
		DownloadURL string `json:"downloadUrl"`
	}
	decodeJSON(t, body, &dl)
	if dl.DownloadURL == "" {
		t.Fatalf("download=%s", body)
	}
}

func TestFilePatch_RenameMove(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	userID := decodeUserID(t, engine, token)
	fileID, objectKey := uploadReady(t, engine, objs, token, "photo.jpg", nil)

	code, body := patchAuth(t, engine, "/api/v1/files/"+fileID, token, map[string]any{
		"name": "renamed.jpg",
	})
	if code != http.StatusOK {
		t.Fatalf("rename status=%d body=%s", code, body)
	}
	var renamed struct {
		Name string `json:"name"`
	}
	decodeJSON(t, body, &renamed)
	if renamed.Name != "renamed.jpg" {
		t.Fatalf("rename=%s", body)
	}

	_, body = postAuth(t, engine, "/api/v1/folders", token, map[string]any{"name": "Docs"})
	var folder struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &folder)

	code, body = patchAuth(t, engine, "/api/v1/files/"+fileID, token, map[string]any{
		"folderId": folder.ID,
	})
	if code != http.StatusOK {
		t.Fatalf("move status=%d body=%s", code, body)
	}
	var moved struct {
		FolderID *string `json:"folderId"`
	}
	decodeJSON(t, body, &moved)
	if moved.FolderID == nil || *moved.FolderID != folder.ID {
		t.Fatalf("move=%s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/browser?folderId="+folder.ID, token)
	if code != http.StatusOK {
		t.Fatalf("browser folder status=%d body=%s", code, body)
	}
	if !strings.Contains(body, "renamed.jpg") {
		t.Fatalf("file not in folder browser: %s", body)
	}

	wantKey := file.ObjectKey(userID, fileID)
	if objectKey != wantKey {
		t.Fatalf("object key changed during test setup")
	}
	code, body = getAuth(t, engine, "/api/v1/files/"+fileID+"/download", token)
	if code != http.StatusOK {
		t.Fatalf("download after patch status=%d body=%s", code, body)
	}
}

func TestFilePatch_DuplicateNameConflict(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	uploadReady(t, engine, objs, token, "a.jpg", nil)
	fileB, _ := uploadReady(t, engine, objs, token, "b.jpg", nil)

	code, body := patchAuth(t, engine, "/api/v1/files/"+fileB, token, map[string]any{
		"name": "a.jpg",
	})
	assertAPIError(t, code, body, http.StatusConflict, "CONFLICT")
}

func TestFilePatch_PendingInvalidState(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        "pending.jpg",
		"size":        64,
		"contentType": "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)

	code, body = patchAuth(t, engine, "/api/v1/files/"+session.FileID, token, map[string]any{
		"name": "nope.jpg",
	})
	assertAPIError(t, code, body, http.StatusConflict, "INVALID_STATE")
}

func TestComplete_WithoutUploadFails(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        "photo.jpg",
		"size":        64,
		"contentType": "image/jpeg",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)

	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", token, nil)
	assertAPIError(t, code, body, http.StatusConflict, "UPLOAD_EXPIRED")
}

func uploadReady(t *testing.T, engine http.Handler, objs *objectstore.Memory, token, name string, folderID *string) (fileID, objectKey string) {
	t.Helper()
	payload := map[string]any{
		"name":        name,
		"size":        128,
		"contentType": "image/jpeg",
	}
	if folderID != nil {
		payload["folderId"] = *folderID
	}
	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, payload)
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objectKey = file.ObjectKey(userID, session.FileID)
	objs.PutObject(objectKey, objectstore.ObjectStat{Size: 128, ContentType: "image/jpeg"})
	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", token, nil)
	if code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	return session.FileID, objectKey
}

func patchAuth(t *testing.T, engine http.Handler, path, token string, payload map[string]any) (int, string) {
	return doJSON(t, engine, http.MethodPatch, path, token, payload)
}

func TestVersioning_ReplaceListDownload(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	// Upload the original file.
	fileID, _ := uploadReady(t, engine, objs, token, "photo.jpg", nil)

	// Replace it with a new version.
	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":          "photo.jpg",
		"size":          256,
		"contentType":   "image/jpeg",
		"replaceFileId": fileID,
	})
	if code != http.StatusCreated {
		t.Fatalf("replace session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	newKey := file.ObjectKey(userID, session.FileID)
	objs.PutObject(newKey, objectstore.ObjectStat{Size: 256, ContentType: "image/jpeg"})
	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", token, nil)
	if code != http.StatusOK {
		t.Fatalf("replace complete status=%d body=%s", code, body)
	}

	// List versions: should have exactly one archived version.
	code, body = getAuth(t, engine, "/api/v1/files/"+fileID+"/versions", token)
	if code != http.StatusOK {
		t.Fatalf("versions status=%d body=%s", code, body)
	}
	var versions struct {
		Versions []struct {
			ID        string `json:"id"`
			SizeBytes int64  `json:"sizeBytes"`
		} `json:"versions"`
	}
	decodeJSON(t, body, &versions)
	if len(versions.Versions) != 1 {
		t.Fatalf("expected 1 version, got %d: %s", len(versions.Versions), body)
	}
	if versions.Versions[0].SizeBytes != 128 {
		t.Fatalf("version size = %d, want 128", versions.Versions[0].SizeBytes)
	}

	// Download the archived version.
	versionID := versions.Versions[0].ID
	code, body = getAuth(t, engine, "/api/v1/files/"+fileID+"/versions/"+versionID+"/download", token)
	if code != http.StatusOK {
		t.Fatalf("version download status=%d body=%s", code, body)
	}
	var dl struct {
		DownloadURL string `json:"downloadUrl"`
	}
	decodeJSON(t, body, &dl)
	if dl.DownloadURL == "" {
		t.Fatalf("version download url empty: %s", body)
	}
}

func TestVersioning_ReplaceNameMismatch(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	fileID, _ := uploadReady(t, engine, objs, token, "photo.jpg", nil)

	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":          "other.jpg",
		"size":          256,
		"contentType":   "image/jpeg",
		"replaceFileId": fileID,
	})
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
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
	return fmt.Sprintf("file-%d@example.com", time.Now().UnixNano())
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
