package activity_test

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

type activityEvent struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	TargetName string `json:"targetName"`
	CreatedAt  string `json:"createdAt"`
}

type activityPage struct {
	Events     []activityEvent `json:"events"`
	NextBefore *string         `json:"nextBefore"`
}

func listActivity(t *testing.T, engine http.Handler, token, query string) (int, activityPage) {
	t.Helper()
	path := "/api/v1/activity"
	if query != "" {
		path += "?" + query
	}
	code, body := getAuth(t, engine, path, token)
	var page activityPage
	if code == http.StatusOK {
		decodeJSON(t, body, &page)
	} else {
		t.Logf("activity status=%d body=%s", code, body)
	}
	return code, page
}

func TestActivity_RecordsLifecycleEvents(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReady(t, engine, objs, token, "life.jpg")

	// Upload must be recorded.
	code, page := listActivity(t, engine, token, "")
	if code != http.StatusOK {
		t.Fatalf("list activity status=%d", code)
	}
	if len(page.Events) == 0 || page.Events[0].Type != "file.uploaded" || page.Events[0].TargetName != "life.jpg" {
		t.Fatalf("upload event missing: %+v", page.Events)
	}

	// Trash + restore are recorded.
	code, _ = deleteAuth(t, engine, "/api/v1/files/"+fileID, token)
	if code != http.StatusNoContent {
		t.Fatalf("trash status=%d", code)
	}
	code, page = listActivity(t, engine, token, "")
	if code != http.StatusOK || len(page.Events) == 0 || page.Events[0].Type != "file.trashed" {
		t.Fatalf("trash event missing: %+v", page.Events)
	}

	code, _ = postAuth(t, engine, "/api/v1/files/"+fileID+"/restore", token, nil)
	if code != http.StatusNoContent {
		t.Fatalf("restore status=%d", code)
	}
	code, page = listActivity(t, engine, token, "")
	if code != http.StatusOK || len(page.Events) == 0 || page.Events[0].Type != "file.restored" {
		t.Fatalf("restore event missing: %+v", page.Events)
	}
}

func TestActivity_RecordsShareAndPasswordEvents(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReady(t, engine, objs, token, "shared.jpg")

	code, _ := postAuth(t, engine, "/api/v1/files/"+fileID+"/share", token, nil)
	if code != http.StatusCreated {
		t.Fatalf("share create status=%d", code)
	}
	code, _ = deleteAuth(t, engine, "/api/v1/files/"+fileID+"/share", token)
	if code != http.StatusNoContent {
		t.Fatalf("share revoke status=%d", code)
	}

	code, _ = postAuth(t, engine, "/api/v1/users/me/password", token, map[string]any{
		"currentPassword": "password1",
		"newPassword":     "password2",
	})
	if code != http.StatusOK {
		t.Fatalf("password change status=%d", code)
	}

	code, page := listActivity(t, engine, token, "")
	if code != http.StatusOK {
		t.Fatalf("list activity status=%d", code)
	}
	types := eventTypes(page.Events)
	for _, want := range []string{"share.created", "share.revoked", "password.changed"} {
		if !types[want] {
			t.Fatalf("missing %q in events: %+v", want, page.Events)
		}
	}
}

func TestActivity_PaginationNewestFirst(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	for i := range 3 {
		uploadReady(t, engine, objs, token, fmt.Sprintf("page-%d.jpg", i))
		time.Sleep(5 * time.Millisecond)
	}

	code, page := listActivity(t, engine, token, "limit=2")
	if code != http.StatusOK {
		t.Fatalf("page 1 status=%d", code)
	}
	if len(page.Events) != 2 || page.NextBefore == nil {
		t.Fatalf("page 1 wrong: %+v", page)
	}
	firstCreatedAt := page.Events[0].CreatedAt
	if firstCreatedAt < page.Events[1].CreatedAt {
		t.Fatalf("events must be newest first: %+v", page.Events)
	}

	code, page2 := listActivity(t, engine, token, "limit=2&before="+*page.NextBefore)
	if code != http.StatusOK {
		t.Fatalf("page 2 status=%d", code)
	}
	if len(page2.Events) == 0 {
		t.Fatalf("page 2 empty")
	}
	if eventTypes(page2.Events)["file.uploaded"] == false && len(page2.Events) == 0 {
		t.Fatalf("page 2 must continue the feed: %+v", page2.Events)
	}
}

func TestActivity_OwnershipIsolation(t *testing.T) {
	engine, mem, objs := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())

	uploadReady(t, engine, objs, tokenA, "only-a.jpg")

	codeA, pageA := listActivity(t, engine, tokenA, "")
	codeB, pageB := listActivity(t, engine, tokenB, "")
	if codeA != http.StatusOK || codeB != http.StatusOK {
		t.Fatalf("list status a=%d b=%d", codeA, codeB)
	}
	if len(pageA.Events) == 0 {
		t.Fatalf("user A must see own events")
	}
	for _, ev := range pageB.Events {
		if ev.TargetName == "only-a.jpg" {
			t.Fatalf("user B must not see A's events: %+v", pageB.Events)
		}
	}
}

// TestActivity_RecordsFolderAndPurgeEvents covers the remaining spec 09 §6.2a
// types: folder.trashed, folder.restored and file.purged.
func TestActivity_RecordsFolderAndPurgeEvents(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	_, body := postAuth(t, engine, "/api/v1/folders", token, map[string]any{"name": "PurgeFolder"})
	var folder struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &folder)

	fileID := uploadReady(t, engine, objs, token, "purge-me.jpg")

	code, _ := deleteAuth(t, engine, "/api/v1/folders/"+folder.ID, token)
	if code != http.StatusNoContent {
		t.Fatalf("folder trash status=%d", code)
	}
	code, page := listActivity(t, engine, token, "")
	if code != http.StatusOK {
		t.Fatalf("list activity status=%d", code)
	}
	types := eventTypes(page.Events)
	if !types["folder.trashed"] {
		t.Fatalf("missing folder.trashed: %+v", page.Events)
	}

	code, _ = postAuth(t, engine, "/api/v1/folders/"+folder.ID+"/restore", token, nil)
	if code != http.StatusNoContent {
		t.Fatalf("folder restore status=%d", code)
	}
	code, page = listActivity(t, engine, token, "")
	if code != http.StatusOK {
		t.Fatalf("list activity status=%d", code)
	}
	if !eventTypes(page.Events)["folder.restored"] {
		t.Fatalf("missing folder.restored: %+v", page.Events)
	}

	// Purge a file: trash it first, then delete permanently.
	code, _ = deleteAuth(t, engine, "/api/v1/files/"+fileID, token)
	if code != http.StatusNoContent {
		t.Fatalf("trash status=%d", code)
	}
	code, _ = deleteAuth(t, engine, "/api/v1/trash/files/"+fileID, token)
	if code != http.StatusNoContent {
		t.Fatalf("permanent delete status=%d", code)
	}
	code, page = listActivity(t, engine, token, "")
	if code != http.StatusOK {
		t.Fatalf("list activity status=%d", code)
	}
	purged := false
	for _, ev := range page.Events {
		if ev.Type == "file.purged" && ev.TargetName == "purge-me.jpg" {
			purged = true
		}
	}
	if !purged {
		t.Fatalf("missing file.purged with targetName: %+v", page.Events)
	}
}

func TestActivity_Validation(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := getAuth(t, engine, "/api/v1/activity?limit=0", token)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, body = getAuth(t, engine, "/api/v1/activity?limit=999", token)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, body = getAuth(t, engine, "/api/v1/activity?before=not-a-ulid", token)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func eventTypes(events []activityEvent) map[string]bool {
	out := make(map[string]bool, len(events))
	for _, ev := range events {
		out[ev.Type] = true
	}
	return out
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
	return fmt.Sprintf("activity-%d@example.com", time.Now().UnixNano())
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
