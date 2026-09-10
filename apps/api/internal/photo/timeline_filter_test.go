package photo_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestPhotos_FilteredTimelineReturnsOnlyRequestedMediaType(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	imageID := uploadReady(t, engine, objs, token, "photo.jpg")
	videoID := uploadReadyVideo(t, engine, objs, token, "clip.mp4")

	code, body := getAuth(t, engine, "/api/v1/photos/timeline/filter?type=video&limit=1", token)
	if code != http.StatusOK {
		t.Fatalf("filtered timeline status=%d body=%s", code, body)
	}
	if !strings.Contains(body, videoID) {
		t.Fatalf("filtered timeline missing video: %s", body)
	}
	if strings.Contains(body, imageID) {
		t.Fatalf("filtered timeline must not include image: %s", body)
	}
}

func TestPhotos_FilteredTimelineRejectsUnsupportedType(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := getAuth(t, engine, "/api/v1/photos/timeline/filter?type=audio", token)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}
