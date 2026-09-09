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

func TestPhotos_ThumbnailURLsFollowUserSettings(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	imageID := uploadReady(t, engine, objs, token, "photo.jpg")
	videoID := uploadReadyVideo(t, engine, objs, token, "clip.mp4")

	code, body := getAuth(t, engine, "/api/v1/photos/timeline", token)
	if code != http.StatusOK {
		t.Fatalf("timeline status=%d body=%s", code, body)
	}
	items := timelineItems(t, body)
	if items[imageID].ThumbnailURL == "" || items[videoID].ThumbnailURL == "" {
		t.Fatalf("expected image and video thumbnails enabled: %s", body)
	}

	code, body = patchAuth(t, engine, "/api/v1/users/me", token, map[string]any{
		"imageThumbnailsEnabled": false,
	})
	if code != http.StatusOK {
		t.Fatalf("disable image thumbnails status=%d body=%s", code, body)
	}
	code, body = getAuth(t, engine, "/api/v1/photos/timeline", token)
	if code != http.StatusOK {
		t.Fatalf("timeline after image setting status=%d body=%s", code, body)
	}
	items = timelineItems(t, body)
	if items[imageID].ThumbnailURL != "" || items[videoID].ThumbnailURL == "" {
		t.Fatalf("image setting was not applied: %s", body)
	}

	code, body = patchAuth(t, engine, "/api/v1/users/me", token, map[string]any{
		"videoThumbnailsEnabled": false,
	})
	if code != http.StatusOK {
		t.Fatalf("disable video thumbnails status=%d body=%s", code, body)
	}
	code, body = getAuth(t, engine, "/api/v1/photos/timeline", token)
	if code != http.StatusOK {
		t.Fatalf("timeline after video setting status=%d body=%s", code, body)
	}
	items = timelineItems(t, body)
	if items[imageID].ThumbnailURL != "" || items[videoID].ThumbnailURL != "" {
		t.Fatalf("video setting was not applied: %s", body)
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

func TestPhotos_AlbumItemCount(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	photoID := uploadReady(t, engine, objs, token, "a.jpg")

	code, body := postAuth(t, engine, "/api/v1/photos/albums", token, map[string]any{"name": "Count"})
	if code != http.StatusCreated {
		t.Fatalf("create album status=%d body=%s", code, body)
	}
	var album struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &album)

	code, body = getAuth(t, engine, "/api/v1/photos/albums", token)
	if code != http.StatusOK {
		t.Fatalf("list albums status=%d body=%s", code, body)
	}
	if !strings.Contains(body, `"itemCount":0`) {
		t.Fatalf("empty album must have itemCount 0: %s", body)
	}

	code, _ = postAuth(t, engine, "/api/v1/photos/albums/"+album.ID+"/items", token, map[string]any{
		"fileIds": []string{photoID},
	})
	if code != http.StatusNoContent {
		t.Fatalf("add item status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums", token)
	if code != http.StatusOK {
		t.Fatalf("list albums status=%d body=%s", code, body)
	}
	if !strings.Contains(body, `"itemCount":1`) {
		t.Fatalf("album must have itemCount 1: %s", body)
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

func createAlbumForTest(t *testing.T, engine http.Handler, token, name string) string {
	t.Helper()
	code, body := postAuth(t, engine, "/api/v1/photos/albums", token, map[string]any{"name": name})
	if code != http.StatusCreated {
		t.Fatalf("create album status=%d body=%s", code, body)
	}
	var album struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &album)
	return album.ID
}

func TestPhotos_AlbumCover_SetRemoveAndAutoFallback(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	firstID := uploadReady(t, engine, objs, token, "first.jpg")
	secondID := uploadReady(t, engine, objs, token, "second.jpg")
	albumID := createAlbumForTest(t, engine, token, "Covers")

	code, _ := postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/items", token, map[string]any{
		"fileIds": []string{firstID, secondID},
	})
	if code != http.StatusNoContent {
		t.Fatalf("add items status=%d", code)
	}

	code, body := getAuth(t, engine, "/api/v1/photos/albums/"+albumID, token)
	if code != http.StatusOK || !strings.Contains(body, `"coverFileId":"`+secondID+`"`) {
		t.Fatalf("auto cover must pick newest item: status=%d body=%s", code, body)
	}

	code, body = postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/cover", token, map[string]any{
		"fileId": firstID,
	})
	if code != http.StatusNoContent {
		t.Fatalf("set cover status=%d body=%s", code, body)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums/"+albumID, token)
	if code != http.StatusOK || !strings.Contains(body, `"coverFileId":"`+firstID+`"`) {
		t.Fatalf("pinned cover not applied: status=%d body=%s", code, body)
	}
	if !strings.Contains(body, `"coverUrl"`) {
		t.Fatalf("image cover must expose coverUrl: %s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums", token)
	if code != http.StatusOK || !strings.Contains(body, `"coverUrl"`) {
		t.Fatalf("list albums must include coverUrl for image cover: %s", body)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/cover", token)
	if code != http.StatusNoContent {
		t.Fatalf("remove cover status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums/"+albumID, token)
	if code != http.StatusOK || strings.Contains(body, `"coverFileId":"`+firstID+`"`) {
		t.Fatalf("remove cover must fall back to auto: %s", body)
	}
}

func TestPhotos_AlbumCover_VideoCoverHasNoURL(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	videoID := uploadReadyVideo(t, engine, objs, token, "clip.mp4")
	albumID := createAlbumForTest(t, engine, token, "Videos")

	code, _ := postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/items", token, map[string]any{
		"fileIds": []string{videoID},
	})
	if code != http.StatusNoContent {
		t.Fatalf("add item status=%d", code)
	}

	code, _ = postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/cover", token, map[string]any{
		"fileId": videoID,
	})
	if code != http.StatusNoContent {
		t.Fatalf("set video cover status=%d", code)
	}

	code, body := getAuth(t, engine, "/api/v1/photos/albums/"+albumID, token)
	if code != http.StatusOK || !strings.Contains(body, `"coverFileId":"`+videoID+`"`) {
		t.Fatalf("video cover not applied: %s", body)
	}
	if strings.Contains(body, `"coverUrl"`) {
		t.Fatalf("video cover must not expose coverUrl: %s", body)
	}
}

func TestPhotos_AlbumCover_FileNotInAlbumRejected(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	outsiderID := uploadReady(t, engine, objs, token, "outsider.jpg")
	pdfID := uploadReadyPDF(t, engine, objs, token, "doc.pdf")
	albumID := createAlbumForTest(t, engine, token, "Strict")

	code, body := postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/cover", token, map[string]any{
		"fileId": outsiderID,
	})
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, body = postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/cover", token, map[string]any{
		"fileId": pdfID,
	})
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, body = postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/cover", token, map[string]any{
		"fileId": "",
	})
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, body = postAuth(t, engine, "/api/v1/photos/albums/01ARZ3NDEKTSV4RRFFQ69G5FAV/cover", token, map[string]any{
		"fileId": outsiderID,
	})
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
}

func TestPhotos_AlbumCover_ClearedWhenItemRemovedOrTrashed(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	itemID := uploadReady(t, engine, objs, token, "pinned.jpg")
	albumID := createAlbumForTest(t, engine, token, "Cleanup")

	code, _ := postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/items", token, map[string]any{
		"fileIds": []string{itemID},
	})
	if code != http.StatusNoContent {
		t.Fatalf("add item status=%d", code)
	}
	code, _ = postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/cover", token, map[string]any{
		"fileId": itemID,
	})
	if code != http.StatusNoContent {
		t.Fatalf("set cover status=%d", code)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/items/"+itemID, token)
	if code != http.StatusNoContent {
		t.Fatalf("remove item status=%d", code)
	}

	code, body := getAuth(t, engine, "/api/v1/photos/albums/"+albumID, token)
	if code != http.StatusOK || strings.Contains(body, `"coverFileId":"`+itemID+`"`) {
		t.Fatalf("removing pinned item must clear cover: %s", body)
	}

	code, _ = postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/items", token, map[string]any{
		"fileIds": []string{itemID},
	})
	if code != http.StatusNoContent {
		t.Fatalf("re-add item status=%d", code)
	}
	code, _ = postAuth(t, engine, "/api/v1/photos/albums/"+albumID+"/cover", token, map[string]any{
		"fileId": itemID,
	})
	if code != http.StatusNoContent {
		t.Fatalf("set cover again status=%d", code)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/files/"+itemID, token)
	if code != http.StatusNoContent {
		t.Fatalf("trash file status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/albums/"+albumID, token)
	if code != http.StatusOK || strings.Contains(body, `"coverFileId":"`+itemID+`"`) {
		t.Fatalf("trashing pinned item must clear cover: %s", body)
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

func uploadReadyVideo(t *testing.T, engine http.Handler, objs *objectstore.Memory, token, name string) string {
	t.Helper()
	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", token, map[string]any{
		"name":        name,
		"size":        128,
		"contentType": "video/mp4",
	})
	if code != http.StatusCreated {
		t.Fatalf("video session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)
	userID := decodeUserID(t, engine, token)
	objectKey := file.ObjectKey(userID, session.FileID)
	objs.PutObject(objectKey, objectstore.ObjectStat{Size: 128, ContentType: "video/mp4"})
	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", token, nil)
	if code != http.StatusOK {
		t.Fatalf("video complete status=%d body=%s", code, body)
	}
	return session.FileID
}

type timelineItemResponse struct {
	ThumbnailURL string `json:"thumbnailUrl"`
}

func timelineItems(t *testing.T, body string) map[string]timelineItemResponse {
	t.Helper()
	var response struct {
		Groups []struct {
			Items []struct {
				ID string `json:"id"`
				timelineItemResponse
			} `json:"items"`
		} `json:"groups"`
	}
	decodeJSON(t, body, &response)
	items := make(map[string]timelineItemResponse)
	for _, group := range response.Groups {
		for _, item := range group.Items {
			items[item.ID] = item.timelineItemResponse
		}
	}
	return items
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
