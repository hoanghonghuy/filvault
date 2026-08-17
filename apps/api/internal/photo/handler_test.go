package photo_test

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

func TestPhotos_Timeline_ExcludesPDF(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	photoID := uploadReady(t, engine, objs, token, "photo.jpg")
	pdfID := uploadReadyPDF(t, engine, objs, token, "doc.pdf")

	code, body := getAuth(t, engine, "/api/v1/photos/timeline", token)
	if code != http.StatusOK {
		t.Fatalf("timeline status=%d body=%s", code, body)
	}
	if !strings.Contains(body, photoID) {
		t.Fatalf("timeline missing photo: %s", body)
	}
	if strings.Contains(body, pdfID) {
		t.Fatalf("timeline must not include pdf: %s", body)
	}
}

func TestPhotos_AlbumCRUD(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuth(t, engine, "/api/v1/photos/albums", token, map[string]any{"name": "Trip"})
	if code != http.StatusCreated {
		t.Fatalf("create album status=%d body=%s", code, body)
	}
	var album struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	decodeJSON(t, body, &album)
	if album.Name != "Trip" || album.ID == "" {
		t.Fatalf("album=%s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums", token)
	if code != http.StatusOK || !strings.Contains(body, "Trip") {
		t.Fatalf("list albums status=%d body=%s", code, body)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums/"+album.ID, token)
	if code != http.StatusOK || !strings.Contains(body, `"items"`) {
		t.Fatalf("get album status=%d body=%s", code, body)
	}

	code, body = patchAuth(t, engine, "/api/v1/photos/albums/"+album.ID, token, map[string]any{"name": "Vacation"})
	if code != http.StatusOK || !strings.Contains(body, "Vacation") {
		t.Fatalf("patch album status=%d body=%s", code, body)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/photos/albums/"+album.ID, token)
	if code != http.StatusNoContent {
		t.Fatalf("delete album status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums/"+album.ID, token)
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
}

func TestPhotos_AlbumItems(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	photoID := uploadReady(t, engine, objs, token, "a.jpg")
	pdfID := uploadReadyPDF(t, engine, objs, token, "b.pdf")

	code, body := postAuth(t, engine, "/api/v1/photos/albums", token, map[string]any{"name": "Best"})
	if code != http.StatusCreated {
		t.Fatalf("create album status=%d body=%s", code, body)
	}
	var album struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &album)

	code, _ = postAuth(t, engine, "/api/v1/photos/albums/"+album.ID+"/items", token, map[string]any{
		"fileIds": []string{photoID},
	})
	if code != http.StatusNoContent {
		t.Fatalf("add item status=%d", code)
	}

	code, _ = postAuth(t, engine, "/api/v1/photos/albums/"+album.ID+"/items", token, map[string]any{
		"fileIds": []string{photoID},
	})
	if code != http.StatusNoContent {
		t.Fatalf("duplicate add must be idempotent status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums/"+album.ID, token)
	if code != http.StatusOK || !strings.Contains(body, photoID) {
		t.Fatalf("album items status=%d body=%s", code, body)
	}

	code, body = postAuth(t, engine, "/api/v1/photos/albums/"+album.ID+"/items", token, map[string]any{
		"fileIds": []string{pdfID},
	})
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, _ = deleteAuth(t, engine, "/api/v1/photos/albums/"+album.ID+"/items/"+photoID, token)
	if code != http.StatusNoContent {
		t.Fatalf("remove item status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums/"+album.ID, token)
	if code != http.StatusOK || strings.Contains(body, photoID) {
		t.Fatalf("album after remove status=%d body=%s", code, body)
	}
}

func TestPhotos_TrashHidden(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	photoID := uploadReady(t, engine, objs, token, "gone.jpg")

	code, body := postAuth(t, engine, "/api/v1/photos/albums", token, map[string]any{"name": "Keep"})
	if code != http.StatusCreated {
		t.Fatalf("create album status=%d body=%s", code, body)
	}
	var album struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &album)

	code, _ = postAuth(t, engine, "/api/v1/photos/albums/"+album.ID+"/items", token, map[string]any{
		"fileIds": []string{photoID},
	})
	if code != http.StatusNoContent {
		t.Fatalf("add item status=%d", code)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/files/"+photoID, token)
	if code != http.StatusNoContent {
		t.Fatalf("soft delete status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/timeline", token)
	if code != http.StatusOK || strings.Contains(body, photoID) {
		t.Fatalf("timeline after trash status=%d body=%s", code, body)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums/"+album.ID, token)
	if code != http.StatusOK || strings.Contains(body, photoID) {
		t.Fatalf("album after trash status=%d body=%s", code, body)
	}
}

func TestPhotos_DuplicateAlbumName(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, _ := postAuth(t, engine, "/api/v1/photos/albums", token, map[string]any{"name": "Same"})
	if code != http.StatusCreated {
		t.Fatalf("create first status=%d", code)
	}

	code, body := postAuth(t, engine, "/api/v1/photos/albums", token, map[string]any{"name": "Same"})
	assertAPIError(t, code, body, http.StatusConflict, "CONFLICT")
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
	return fmt.Sprintf("photo-%d@example.com", time.Now().UnixNano())
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

func patchAuth(t *testing.T, engine http.Handler, path, token string, payload map[string]any) (int, string) {
	return doJSON(t, engine, http.MethodPatch, path, token, payload)
}

func getAuth(t *testing.T, engine http.Handler, path, token string) (int, string) {
	return doJSON(t, engine, http.MethodGet, path, token, nil)
}

func deleteAuth(t *testing.T, engine http.Handler, path, token string) (int, string) {
	return doJSON(t, engine, http.MethodDelete, path, token, nil)
}
