package storage_test

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

func TestStorage_ReturnsUsage(t *testing.T) {
	engine, mem, objs, pool := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	uploadReady(t, engine, objs, token, "a.jpg")

	code, body := getAuth(t, engine, "/api/v1/storage", token)
	if code != http.StatusOK {
		t.Fatalf("storage status=%d body=%s", code, body)
	}
	var out struct {
		UsedBytes  int64 `json:"usedBytes"`
		QuotaBytes int64 `json:"quotaBytes"`
	}
	decodeJSON(t, body, &out)
	if out.QuotaBytes != config.DefaultStorageQuotaBytes {
		t.Fatalf("quota=%d", out.QuotaBytes)
	}
	if out.UsedBytes != 128 {
		t.Fatalf("used=%d want 128", out.UsedBytes)
	}

	userID := decodeUserID(t, engine, token)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `UPDATE users SET storage_used = storage_quota WHERE id = $1`, userID); err != nil {
		t.Fatalf("fill quota: %v", err)
	}

	code, body = postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        "full.jpg",
		"size":        1,
		"contentType": "image/jpeg",
	})
	assertAPIError(t, code, body, http.StatusConflict, "QUOTA_EXCEEDED")
}

func uploadReady(t *testing.T, engine http.Handler, objs *objectstore.Memory, token, name string) {
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
}

func newEngine(t *testing.T) (*gin.Engine, *mailer.Memory, *objectstore.Memory, *pgxpool.Pool) {
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
	return engine, mem, objs, pool
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
	_, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       email,
		"password":    "password1",
		"displayName": "User",
		"inviteCode":  "secret-invite",
	})
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
	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
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
	return fmt.Sprintf("storage-%d@example.com", time.Now().UnixNano())
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
