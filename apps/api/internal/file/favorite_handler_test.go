package file_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestFavorites_PutListDelete(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID, _ := uploadReady(t, engine, objs, token, "fav.jpg", nil)
	otherID, _ := uploadReady(t, engine, objs, token, "plain.jpg", nil)

	code, _ := putAuth(t, engine, "/api/v1/files/"+fileID+"/favorite", token)
	if code != http.StatusNoContent {
		t.Fatalf("put favorite status=%d", code)
	}
	// Idempotent PUT.
	code, _ = putAuth(t, engine, "/api/v1/files/"+fileID+"/favorite", token)
	if code != http.StatusNoContent {
		t.Fatalf("idempotent put favorite status=%d", code)
	}

	code, body := getAuth(t, engine, "/api/v1/files/favorites", token)
	if code != http.StatusOK {
		t.Fatalf("list favorites status=%d body=%s", code, body)
	}
	if !strings.Contains(body, fileID) || strings.Contains(body, otherID) {
		t.Fatalf("favorites list wrong: %s", body)
	}
	if !strings.Contains(body, `"favoritedAt"`) {
		t.Fatalf("favorites must include favoritedAt: %s", body)
	}

	code, _ = deleteAuth(t, engine, "/api/v1/files/"+fileID+"/favorite", token)
	if code != http.StatusNoContent {
		t.Fatalf("delete favorite status=%d", code)
	}
	// Idempotent DELETE.
	code, _ = deleteAuth(t, engine, "/api/v1/files/"+fileID+"/favorite", token)
	if code != http.StatusNoContent {
		t.Fatalf("idempotent delete favorite status=%d", code)
	}

	code, body = getAuth(t, engine, "/api/v1/files/favorites", token)
	if code != http.StatusOK || strings.Contains(body, fileID) {
		t.Fatalf("favorites after delete: %s", body)
	}
}

func TestFavorites_NewestFirstAndLimit(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	var order []string
	for i := range 3 {
		id, _ := uploadReady(t, engine, objs, token, fmt.Sprintf("f%d.jpg", i), nil)
		order = append(order, id)
		code, body := putAuth(t, engine, "/api/v1/files/"+id+"/favorite", token)
		if code != http.StatusNoContent {
			t.Fatalf("put %d status=%d body=%s", i, code, body)
		}
		time.Sleep(5 * time.Millisecond)
	}

	code, body := getAuth(t, engine, "/api/v1/files/favorites?limit=2", token)
	if code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", code, body)
	}
	if !strings.Contains(body, order[2]) || !strings.Contains(body, order[1]) {
		t.Fatalf("newest first violated: %s", body)
	}
	if strings.Contains(body, order[0]) {
		t.Fatalf("limit not applied: %s", body)
	}
}

func TestFavorites_TrashHidesAndPermanentDeleteRemovesRow(t *testing.T) {
	engine, mem, objs := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	trashedID, _ := uploadReady(t, engine, objs, token, "gone.jpg", nil)
	purgedID, _ := uploadReady(t, engine, objs, token, "purged.jpg", nil)

	for _, id := range []string{trashedID, purgedID} {
		code, body := putAuth(t, engine, "/api/v1/files/"+id+"/favorite", token)
		if code != http.StatusNoContent {
			t.Fatalf("put %s status=%d body=%s", id, code, body)
		}
	}

	code, _ := deleteAuth(t, engine, "/api/v1/files/"+trashedID, token)
	if code != http.StatusNoContent {
		t.Fatalf("soft delete status=%d", code)
	}

	code, body := getAuth(t, engine, "/api/v1/files/favorites", token)
	if code != http.StatusOK || strings.Contains(body, trashedID) {
		t.Fatalf("trashed file must be hidden from favorites: %s", body)
	}

	// Permanent delete removes the row (FK cascade).
	code, _ = deleteAuth(t, engine, "/api/v1/files/"+purgedID, token)
	if code != http.StatusNoContent {
		t.Fatalf("soft delete second status=%d", code)
	}
	code, body = deleteAuth(t, engine, "/api/v1/trash/files/"+purgedID, token)
	if code != http.StatusNoContent {
		t.Fatalf("purge status=%d body=%s", code, body)
	}

	code, body = getAuth(t, engine, "/api/v1/files/favorites", token)
	if code != http.StatusOK || strings.Contains(body, purgedID) {
		t.Fatalf("purged file row must be removed from favorites: %s", body)
	}
}

func TestFavorites_UnknownFileNotFound(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := putAuth(t, engine, "/api/v1/files/01ARZ3NDEKTSV4RRFFQ69G5FAV/favorite", token)
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")

	code, body = deleteAuth(t, engine, "/api/v1/files/01ARZ3NDEKTSV4RRFFQ69G5FAV/favorite", token)
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
}

func putAuth(t *testing.T, engine http.Handler, path, token string) (int, string) {
	return doJSON(t, engine, http.MethodPut, path, token, nil)
}

func deleteAuth(t *testing.T, engine http.Handler, path, token string) (int, string) {
	return doJSON(t, engine, http.MethodDelete, path, token, nil)
}
