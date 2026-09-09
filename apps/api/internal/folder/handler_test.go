package folder_test

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
	"filvault/internal/platform/config"
	"filvault/internal/platform/mailer"
	"filvault/internal/platform/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestFolders_UnverifiedForbidden(t *testing.T) {
	engine, _ := newEngine(t)
	email := uniqueEmail()
	token := registerOnly(t, engine, email)

	code, body := postAuth(t, engine, "/api/v1/folders", token, map[string]any{
		"name": "Docs",
	})
	assertAPIError(t, code, body, http.StatusForbidden, "EMAIL_NOT_VERIFIED")
}

func TestFolders_CreateBrowserRenameMoveDelete(t *testing.T) {
	engine, mem := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuth(t, engine, "/api/v1/folders", token, map[string]any{
		"name": "Documents",
	})
	if code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", code, body)
	}
	var doc folderResp
	decodeJSON(t, body, &doc)
	if doc.Name != "Documents" || doc.ParentID != nil {
		t.Fatalf("folder=%+v", doc)
	}

	code, body = getAuth(t, engine, "/api/v1/browser", token)
	if code != http.StatusOK {
		t.Fatalf("browser status=%d body=%s", code, body)
	}
	var root browserResp
	decodeJSON(t, body, &root)
	if len(root.Folders) != 1 || root.Folders[0].ID != doc.ID {
		t.Fatalf("browser root=%s", body)
	}
	if len(root.Files) != 0 || root.Folder != nil {
		t.Fatalf("expected empty files and null folder at root")
	}

	code, body = postAuth(t, engine, "/api/v1/folders", token, map[string]any{
		"name": "Documents",
	})
	assertAPIError(t, code, body, http.StatusConflict, "CONFLICT")

	code, body = postAuth(t, engine, "/api/v1/folders", token, map[string]any{
		"name":     "Work",
		"parentId": doc.ID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create child status=%d body=%s", code, body)
	}
	var work folderResp
	decodeJSON(t, body, &work)
	if work.ParentID == nil || *work.ParentID != doc.ID {
		t.Fatalf("work parent=%v", work.ParentID)
	}

	code, out := deleteAuth(t, engine, "/api/v1/folders/"+doc.ID, token)
	assertAPIError(t, code, out, http.StatusConflict, "CONFLICT")

	code, body = patchAuth(t, engine, "/api/v1/folders/"+work.ID, token, map[string]any{
		"name": "Work renamed",
	})
	if code != http.StatusOK {
		t.Fatalf("rename status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &work)
	if work.Name != "Work renamed" {
		t.Fatalf("rename failed: %s", body)
	}

	code, body = postAuth(t, engine, "/api/v1/folders", token, map[string]any{
		"name": "Archive",
	})
	if code != http.StatusCreated {
		t.Fatalf("archive create status=%d body=%s", code, body)
	}
	var archive folderResp
	decodeJSON(t, body, &archive)

	code, body = patchAuth(t, engine, "/api/v1/folders/"+work.ID, token, map[string]any{
		"parentId": archive.ID,
	})
	if code != http.StatusOK {
		t.Fatalf("move status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &work)
	if work.ParentID == nil || *work.ParentID != archive.ID {
		t.Fatalf("move failed: %s", body)
	}

	code, body = patchAuth(t, engine, "/api/v1/folders/"+archive.ID, token, map[string]any{
		"parentId": work.ID,
	})
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, _ = deleteAuth(t, engine, "/api/v1/folders/"+work.ID, token)
	if code != http.StatusNoContent {
		t.Fatalf("delete leaf status=%d", code)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/folders/"+archive.ID, token)
	if code != http.StatusNoContent {
		t.Fatalf("delete empty archive status=%d", code)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/folders/"+doc.ID, token)
	if code != http.StatusNoContent {
		t.Fatalf("delete documents after child removed status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/browser", token)
	if code != http.StatusOK {
		t.Fatalf("browser after delete status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &root)
	if len(root.Folders) != 0 {
		t.Fatalf("expected no folders, got %s", body)
	}
}

func TestFolders_GetOrCreate(t *testing.T) {
	engine, mem := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	// First call creates the folder.
	code, body := postAuth(t, engine, "/api/v1/folders/get-or-create", token, map[string]any{
		"name": "Photos",
	})
	if code != http.StatusCreated {
		t.Fatalf("get-or-create create status=%d body=%s", code, body)
	}
	var f folderResp
	decodeJSON(t, body, &f)
	if f.Name != "Photos" || f.ParentID != nil {
		t.Fatalf("folder=%+v", f)
	}

	// Second call with the same name returns the existing folder (200).
	code, body = postAuth(t, engine, "/api/v1/folders/get-or-create", token, map[string]any{
		"name": "Photos",
	})
	if code != http.StatusOK {
		t.Fatalf("get-or-create existing status=%d body=%s", code, body)
	}
	var f2 folderResp
	decodeJSON(t, body, &f2)
	if f2.ID != f.ID {
		t.Fatalf("expected same folder id, got %s vs %s", f2.ID, f.ID)
	}

	// Nested under a parent.
	code, body = postAuth(t, engine, "/api/v1/folders/get-or-create", token, map[string]any{
		"name":     "2026",
		"parentId": f.ID,
	})
	if code != http.StatusCreated {
		t.Fatalf("get-or-create child status=%d body=%s", code, body)
	}
	var child folderResp
	decodeJSON(t, body, &child)
	if child.ParentID == nil || *child.ParentID != f.ID {
		t.Fatalf("child parent=%v", child.ParentID)
	}

	// Empty name is rejected.
	code, body = postAuth(t, engine, "/api/v1/folders/get-or-create", token, map[string]any{
		"name": "   ",
	})
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func TestFolders_InvalidULID(t *testing.T) {
	engine, mem := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := getAuth(t, engine, "/api/v1/folders/not-a-ulid", token)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func TestFolders_GetNotFoundForOtherUser(t *testing.T) {
	engine, mem := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/folders", tokenA, map[string]any{
		"name": "Private",
	})
	var f folderResp
	decodeJSON(t, body, &f)

	code, out := getAuth(t, engine, "/api/v1/folders/"+f.ID, tokenB)
	assertAPIError(t, code, out, http.StatusNotFound, "NOT_FOUND")
}

type folderResp struct {
	ID        string  `json:"id"`
	ParentID  *string `json:"parentId"`
	Name      string  `json:"name"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

type browserResp struct {
	Folder     *folderResp  `json:"folder"`
	Breadcrumb []folderResp `json:"breadcrumb"`
	Folders    []folderResp `json:"folders"`
	Files      []any        `json:"files"`
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

func newEngine(t *testing.T) (*gin.Engine, *mailer.Memory) {
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
	engine := app.NewWithMailer(config.Config{
		InviteCode:                "secret-invite",
		JWTSecret:                 "test-jwt-secret-not-for-prod",
		DefaultTrashAutoDelete:    false,
		DefaultTrashRetentionDays: config.DefaultTrashRetentionDays,
		MetadataStore:             "postgres",
	}, pool, mem)
	return engine, mem
}

func uniqueEmail() string {
	return fmt.Sprintf("folder-%d@example.com", time.Now().UnixNano())
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

func patchAuth(t *testing.T, engine http.Handler, path, token string, payload map[string]any) (int, string) {
	return doJSON(t, engine, http.MethodPatch, path, token, payload)
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
