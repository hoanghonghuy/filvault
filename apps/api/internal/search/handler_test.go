package search_test

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

func TestSearch_FindsFilesAndFolders(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/folders", token, map[string]any{"name": "Holiday Photos"})
	var folder struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &folder)
	uploadReady(t, engine, objs, token, "holiday-beach.jpg")

	code, body := getAuth(t, engine, "/api/v1/search?q=holiday", token)
	if code != http.StatusOK {
		t.Fatalf("search status=%d body=%s", code, body)
	}
	if !strings.Contains(body, "Holiday Photos") {
		t.Fatalf("search missing folder: %s", body)
	}
	if !strings.Contains(body, "holiday-beach.jpg") {
		t.Fatalf("search missing file: %s", body)
	}
}

func TestSearch_EmptyQueryRejected(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := getAuth(t, engine, "/api/v1/search?q=", token)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func TestSearch_ExcludesTrash(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReadyPDF(t, engine, objs, token, "secret-report.pdf")

	code, _ := deleteAuth(t, engine, "/api/v1/files/"+fileID, token)
	if code != http.StatusNoContent {
		t.Fatalf("delete status=%d", code)
	}

	code, body := getAuth(t, engine, "/api/v1/search?q=secret", token)
	if code != http.StatusOK {
		t.Fatalf("search status=%d body=%s", code, body)
	}
	if strings.Contains(body, fileID) {
		t.Fatalf("trashed file must not appear: %s", body)
	}
}

func uploadReadyPDF(t *testing.T, engine http.Handler, objs *objectstore.Memory, token, name string) string {
	t.Helper()
	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        name,
		"size":        64,
		"contentType": "application/pdf",
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
	objs.PutObject(objectKey, objectstore.ObjectStat{Size: 64, ContentType: "application/pdf"})
	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", token, nil)
	if code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	return session.FileID
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
	return fmt.Sprintf("search-%d@example.com", time.Now().UnixNano())
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
	decodeJSON(t, body, &payload)
	if payload.Error.Code != wantCode {
		t.Fatalf("code=%q want=%q body=%s", payload.Error.Code, wantCode, body)
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
