package trash_test

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

	"filnest/internal/app"
	"filnest/internal/platform/config"
	"filnest/internal/platform/mailer"
	"filnest/internal/platform/objectstore"
	"filnest/internal/platform/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestTrash_SoftDeleteRestorePermanent(t *testing.T) {
	engine, mem, objs, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReady(t, engine, objs, token, "photo.jpg")

	code, _ := deleteAuth(t, engine, "/api/v1/files/"+fileID, token)
	if code != http.StatusNoContent {
		t.Fatalf("soft delete status=%d", code)
	}

	code, body := getAuth(t, engine, "/api/v1/browser", token)
	if code != http.StatusOK {
		t.Fatalf("browser after delete status=%d", code)
	}
	if contains(body, "photo.jpg") {
		t.Fatalf("file still in browser: %s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/trash", token)
	if code != http.StatusOK {
		t.Fatalf("trash list status=%d body=%s", code, body)
	}
	if !contains(body, "photo.jpg") {
		t.Fatalf("trash missing file: %s", body)
	}

	code, _ = postAuth(t, engine, "/api/v1/files/"+fileID+"/restore", token, nil)
	if code != http.StatusNoContent {
		t.Fatalf("restore status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/browser", token)
	if code != http.StatusOK || !contains(body, "photo.jpg") {
		t.Fatalf("browser after restore: %s", body)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/files/"+fileID, token)
	if code != http.StatusNoContent {
		t.Fatalf("re-delete status=%d", code)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/trash/files/"+fileID, token)
	if code != http.StatusNoContent {
		t.Fatalf("permanent delete status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/trash", token)
	if code != http.StatusOK || contains(body, fileID) {
		t.Fatalf("trash after permanent: %s", body)
	}
}

func TestTrash_AutoCleanup(t *testing.T) {
	engine, mem, objs, pool := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReady(t, engine, objs, token, "old.jpg")

	code, _ := deleteAuth(t, engine, "/api/v1/files/"+fileID, token)
	if code != http.StatusNoContent {
		t.Fatalf("delete status=%d", code)
	}

	code, body := patchAuth(t, engine, "/api/v1/users/me", token, map[string]any{
		"trashAutoDeleteEnabled": true,
		"trashRetentionDays":     1,
	})
	if code != http.StatusOK {
		t.Fatalf("patch me status=%d body=%s", code, body)
	}

	ctx := context.Background()
	if _, err := pool.Exec(ctx, `
		UPDATE files SET deleted_at = now() - interval '2 days'
		WHERE id = $1
	`, fileID); err != nil {
		t.Fatalf("backdate deleted_at: %v", err)
	}

	trashSvc := app.NewTrashService(pool, objs)
	if err := trashSvc.RunAutoCleanup(ctx); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	code, body = getAuth(t, engine, "/api/v1/trash", token)
	if code != http.StatusOK || contains(body, fileID) {
		t.Fatalf("trash should be empty after cleanup: %s", body)
	}
}

func TestTrash_FolderSoftDeleteRestore(t *testing.T) {
	engine, mem, _, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/folders", token, map[string]any{"name": "Archive"})
	var folder struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &folder)

	code, _ := deleteAuth(t, engine, "/api/v1/folders/"+folder.ID, token)
	if code != http.StatusNoContent {
		t.Fatalf("folder delete status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/trash", token)
	if code != http.StatusOK || !contains(body, "Archive") {
		t.Fatalf("trash folder: %s", body)
	}

	code, _ = postAuth(t, engine, "/api/v1/folders/"+folder.ID+"/restore", token, nil)
	if code != http.StatusNoContent {
		t.Fatalf("folder restore status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/browser", token)
	if code != http.StatusOK || !contains(body, "Archive") {
		t.Fatalf("browser after folder restore: %s", body)
	}
}

func newEngine(t *testing.T) (*gin.Engine, *mailer.Memory, *objectstore.Memory, *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	url := os.Getenv("FILNEST_DATABASE_URL")
	if url == "" {
		t.Fatal("FILNEST_DATABASE_URL is required")
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
	return engine, mem, objs, pool
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
	code, body = getAuth(t, engine, "/api/v1/users/me", token)
	if code != http.StatusOK {
		t.Fatalf("me status=%d", code)
	}
	var me struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &me)
	objs.PutObject("users/"+me.ID+"/files/"+session.FileID, objectstore.ObjectStat{Size: 128, ContentType: "image/jpeg"})
	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", token, nil)
	if code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	return session.FileID
}

func registerVerified(t *testing.T, engine http.Handler, mem *mailer.Memory, email string) string {
	t.Helper()
	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email": email, "password": "password1", "displayName": "User", "inviteCode": "secret-invite",
	})
	if code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", code, body)
	}
	postJSON(t, engine, "/api/v1/auth/resend-verification", map[string]any{"email": email})
	postJSON(t, engine, "/api/v1/auth/verify-email", map[string]any{"email": email, "code": mem.LastCode(email)})
	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{"email": email, "password": "password1"})
	var s struct {
		AccessToken string `json:"accessToken"`
	}
	decodeJSON(t, body, &s)
	return s.AccessToken
}

func uniqueEmail() string {
	return fmt.Sprintf("trash-%d@example.com", time.Now().UnixNano())
}

func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}

func decodeJSON(t *testing.T, body string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		t.Fatalf("json: %v body=%s", err, body)
	}
}

func postJSON(t *testing.T, engine http.Handler, path string, payload map[string]any) (int, string) {
	return doJSON(t, engine, http.MethodPost, path, "", payload)
}

func postAuth(t *testing.T, engine http.Handler, path, token string, payload map[string]any) (int, string) {
	return doJSON(t, engine, http.MethodPost, path, token, payload)
}

func patchAuth(t *testing.T, engine http.Handler, path, token string, payload map[string]any) (int, string) {
	return doJSON(t, engine, http.MethodPatch, path, token, payload)
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
