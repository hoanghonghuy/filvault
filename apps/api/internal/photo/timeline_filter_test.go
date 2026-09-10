package photo_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"filvault/internal/photo"
)

func TestPhotos_FilteredTimelineReturnsOnlyRequestedMediaTypeAndFavoriteState(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	imageID := uploadReady(t, engine, objs, token, "photo.jpg")
	videoID := uploadReadyVideo(t, engine, objs, token, "clip.mp4")

	code, body := doJSON(t, engine, http.MethodPut, "/api/v1/files/"+videoID+"/favorite", token, nil)
	if code != http.StatusNoContent {
		t.Fatalf("favorite video status=%d body=%s", code, body)
	}

	code, body = getAuth(t, engine, "/api/v1/photos/timeline/filter?type=video&limit=1", token)
	if code != http.StatusOK {
		t.Fatalf("filtered timeline status=%d body=%s", code, body)
	}
	if !strings.Contains(body, videoID) {
		t.Fatalf("filtered timeline missing video: %s", body)
	}
	if strings.Contains(body, imageID) {
		t.Fatalf("filtered timeline must not include image: %s", body)
	}
	if !strings.Contains(body, `"isFavorite":true`) {
		t.Fatalf("filtered timeline must include authoritative favorite state: %s", body)
	}
}

func TestPhotos_FilteredTimelineRejectsUnsupportedTypeAndMalformedCursor(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := getAuth(t, engine, "/api/v1/photos/timeline/filter?type=audio", token)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, body = getAuth(t, engine, "/api/v1/photos/timeline/filter?type=video&before=not-a-cursor", token)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func TestPhotos_FilteredTimelineCompositeCursorRoundTrips(t *testing.T) {
	createdAt := time.Date(2026, 9, 10, 20, 0, 0, 123456789, time.UTC)
	item := photo.TimelineItem{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", CreatedAt: createdAt}

	raw := photo.FormatTimelineCursor(item)
	cursor, err := photo.ParseTimelineCursor(raw)
	if err != nil {
		t.Fatalf("parse cursor: %v", err)
	}
	if cursor == nil || cursor.ID != item.ID || !cursor.CreatedAt.Equal(createdAt) {
		t.Fatalf("cursor round trip mismatch: raw=%q cursor=%+v", raw, cursor)
	}
}
