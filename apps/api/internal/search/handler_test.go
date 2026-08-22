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

type searchPayload struct {
	Folders []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"folders"`
	Files []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		MimeType  string `json:"mimeType"`
		SizeBytes int64  `json:"sizeBytes"`
	} `json:"files"`
}

func searchNames(p searchPayload) []string {
	names := make([]string, 0, len(p.Files))
	for _, f := range p.Files {
		names = append(names, f.Name)
	}
	return names
}

func containsString(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func TestSearch_TypeFilter(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	uploadReady(t, engine, objs, token, "trip-photo.jpg")
	uploadReadyPDF(t, engine, objs, token, "trip-notes.pdf")

	code, body := getAuth(t, engine, "/api/v1/search?q=trip&type=image", token)
	if code != http.StatusOK {
		t.Fatalf("image search status=%d body=%s", code, body)
	}
	var p searchPayload
	decodeJSON(t, body, &p)
	if len(p.Folders) != 0 || len(p.Files) != 1 || p.Files[0].Name != "trip-photo.jpg" {
		t.Fatalf("image filter mismatch: %s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/search?q=trip&type=document", token)
	if code != http.StatusOK {
		t.Fatalf("document search status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &p)
	if len(p.Files) != 1 || p.Files[0].Name != "trip-notes.pdf" {
		t.Fatalf("document filter mismatch: %s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/search?q=trip&type=all", token)
	if code != http.StatusOK {
		t.Fatalf("all search status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &p)
	if len(p.Files) != 2 {
		t.Fatalf("all filter mismatch: %s", body)
	}
}

func TestSearch_TypeFolderReturnsFoldersOnly(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	postAuth(t, engine, "/api/v1/folders", token, map[string]any{"name": "trip-2026"})
	uploadReady(t, engine, objs, token, "trip-photo.jpg")

	code, body := getAuth(t, engine, "/api/v1/search?q=trip&type=folder", token)
	if code != http.StatusOK {
		t.Fatalf("status=%d body=%s", code, body)
	}
	var p searchPayload
	decodeJSON(t, body, &p)
	if len(p.Folders) != 1 || p.Folders[0].Name != "trip-2026" {
		t.Fatalf("folder filter mismatch: %s", body)
	}
	if len(p.Files) != 0 {
		t.Fatalf("type=folder must not return files: %s", body)
	}
}

func TestSearch_FolderIdScopesToSubtree(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	code, body := postAuth(t, engine, "/api/v1/folders", token, map[string]any{"name": "parent"})
	if code != http.StatusCreated {
		t.Fatalf("create parent status=%d body=%s", code, body)
	}
	var parent struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &parent)

	code, body = postAuth(t, engine, "/api/v1/folders", token, map[string]any{"name": "child", "parentId": parent.ID})
	if code != http.StatusCreated {
		t.Fatalf("create child status=%d body=%s", code, body)
	}
	var child struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &child)

	code, body = postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{"name": "elsewhere.txt", "size": 16, "contentType": "text/plain"})
	if code != http.StatusCreated {
		t.Fatalf("create root session status=%d body=%s", code, body)
	}

	code, body = postAuth(t, engine, fmt.Sprintf("/api/v1/files/upload-sessions"), token, map[string]any{"name": "inside.txt", "size": 16, "contentType": "text/plain", "folderId": child.ID})
	if code != http.StatusCreated {
		t.Fatalf("create nested session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objKey := file.ObjectKey(userID, session.FileID)
	objs.PutObject(objKey, objectstore.ObjectStat{Size: 16, ContentType: "text/plain"})
	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", token, nil)
	if code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", code, body)
	}

	code, body = getAuth(t, engine, "/api/v1/search?q=.txt&folderId="+parent.ID, token)
	if code != http.StatusOK {
		t.Fatalf("scoped status=%d body=%s", code, body)
	}
	var p searchPayload
	decodeJSON(t, body, &p)
	names := searchNames(p)
	if len(names) != 1 || !containsString(names, "inside.txt") {
		t.Fatalf("scoped subtree mismatch: %v %s", names, body)
	}
}

func TestSearch_ValidationErrors(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	cases := []struct {
		name string
		qs   string
	}{
		{"unknown type", "?q=x&type=audio"},
		{"unknown sort", "?q=x&sort=sizee"},
		{"unknown order", "?q=x&order=sideways"},
		{"bad limit zero", "?q=x&limit=0"},
		{"bad limit over cap", "?q=x&limit=51"},
		{"bad ulid folderId", "?q=x&folderId=nope"},
		{"bad from date", "?q=x&from=2026-13-40"},
		{"from after to", "?q=x&from=2026-05-02&to=2026-05-01"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, body := getAuth(t, engine, "/api/v1/search"+tc.qs, token)
			assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
		})
	}
}

func TestSearch_FolderIdUnknownNotFound(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	foreign := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	code, body := getAuth(t, engine, "/api/v1/search?q=x&folderId="+foreign, token)
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
}

func TestSearch_DateRange(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	uploadSized(t, engine, objs, token, "dated-report.pdf", "application/pdf", 64)

	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")

	code, body := getAuth(t, engine, "/api/v1/search?q=dated&from="+yesterday+"&to="+tomorrow, token)
	if code != http.StatusOK {
		t.Fatalf("in-range status=%d body=%s", code, body)
	}
	var p searchPayload
	decodeJSON(t, body, &p)
	if len(p.Files) != 1 || p.Files[0].Name != "dated-report.pdf" {
		t.Fatalf("in-range mismatch: %s", body)
	}

	future := now.AddDate(0, 0, 30).Format("2006-01-02")
	code, body = getAuth(t, engine, "/api/v1/search?q=dated&from="+future, token)
	if code != http.StatusOK {
		t.Fatalf("future-from status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &p)
	if len(p.Files) != 0 {
		t.Fatalf("future-from must exclude existing file: %s", body)
	}
}

func TestSearch_SortSize(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	uploadSized(t, engine, objs, token, "small.txt", "text/plain", 32)
	uploadSized(t, engine, objs, token, "big.zip", "application/zip", 256)

	code, body := getAuth(t, engine, "/api/v1/search?q=.&sort=size&order=desc", token)
	if code != http.StatusOK {
		t.Fatalf("desc status=%d body=%s", code, body)
	}
	var p searchPayload
	decodeJSON(t, body, &p)
	if len(p.Files) != 2 || p.Files[0].Name != "big.zip" || p.Files[1].Name != "small.txt" {
		t.Fatalf("desc order mismatch: %+v", searchNames(p))
	}

	code, body = getAuth(t, engine, "/api/v1/search?q=.&sort=size&order=asc", token)
	if code != http.StatusOK {
		t.Fatalf("asc status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &p)
	if len(p.Files) != 2 || p.Files[0].Name != "small.txt" || p.Files[1].Name != "big.zip" {
		t.Fatalf("asc order mismatch: %+v", searchNames(p))
	}
}

func TestSearch_SortDateDefaultDesc(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	olderID := uploadSized(t, engine, objs, token, "older.txt", "text/plain", 8)
	time.Sleep(50 * time.Millisecond)
	newerID := uploadSized(t, engine, objs, token, "newer.txt", "text/plain", 8)

	code, body := getAuth(t, engine, "/api/v1/search?q=.txt&sort=date", token)
	if code != http.StatusOK {
		t.Fatalf("status=%d body=%s", code, body)
	}
	var p searchPayload
	decodeJSON(t, body, &p)
	if len(p.Files) != 2 || p.Files[0].ID != newerID || p.Files[1].ID != olderID {
		t.Fatalf("date desc mismatch: %+v", searchNames(p))
	}
}

func uploadSized(t *testing.T, engine http.Handler, objs *objectstore.Memory, token, name, contentType string, size int64) string {
	t.Helper()
	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        name,
		"size":        size,
		"contentType": contentType,
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
	objs.PutObject(objectKey, objectstore.ObjectStat{Size: size, ContentType: contentType})
	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", token, nil)
	if code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", code, body)
	}
	return session.FileID
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
