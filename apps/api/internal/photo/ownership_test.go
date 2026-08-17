package photo_test

import (
	"net/http"
	"testing"
)

func TestPhoto_AlbumOtherUserNotFound(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuth(t, engine, "/api/v1/photos/albums", tokenA, map[string]any{"name": "Private"})
	if code != http.StatusCreated {
		t.Fatalf("create album status=%d body=%s", code, body)
	}
	var album struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &album)

	cases := []struct {
		method string
		path   string
		body   map[string]any
	}{
		{http.MethodGet, "/api/v1/photos/albums/" + album.ID, nil},
		{http.MethodPatch, "/api/v1/photos/albums/" + album.ID, map[string]any{"name": "Hacked"}},
		{http.MethodDelete, "/api/v1/photos/albums/" + album.ID, nil},
	}
	for _, tc := range cases {
		var code int
		var out string
		switch tc.method {
		case http.MethodGet:
			code, out = getAuth(t, engine, tc.path, tokenB)
		case http.MethodPatch:
			code, out = patchAuth(t, engine, tc.path, tokenB, tc.body)
		case http.MethodDelete:
			code, out = deleteAuth(t, engine, tc.path, tokenB)
		}
		assertAPIError(t, code, out, http.StatusNotFound, "NOT_FOUND")
	}
}
