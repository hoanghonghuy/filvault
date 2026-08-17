package file_test

import (
	"net/http"
	"testing"

	"filnest/internal/file"
	"filnest/internal/platform/objectstore"
)

func TestFile_OtherUserCannotAccess(t *testing.T) {
	engine, mem, objs := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())
	fileID, _ := uploadReady(t, engine, objs, tokenA, "private.jpg", nil)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/files/" + fileID},
		{http.MethodGet, "/api/v1/files/" + fileID + "/download"},
	}
	for _, tc := range cases {
		code, body := doJSON(t, engine, tc.method, tc.path, tokenB, nil)
		assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
	}

	code, body := patchAuth(t, engine, "/api/v1/files/"+fileID, tokenB, map[string]any{"name": "stolen.jpg"})
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")

	code, body = doJSON(t, engine, http.MethodDelete, "/api/v1/files/"+fileID, tokenB, nil)
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
}

func TestFile_OtherUserCannotComplete(t *testing.T) {
	engine, mem, objs := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail())
	tokenB := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuth(t, engine, "/api/v1/files/upload-sessions", tokenA, map[string]any{
		"name":        "pending.jpg",
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
	userID := decodeUserID(t, engine, tokenA)
	objectKey := file.ObjectKey(userID, session.FileID)
	objs.PutObject(objectKey, objectstore.ObjectStat{Size: 128, ContentType: "image/jpeg"})

	code, body = postAuth(t, engine, "/api/v1/files/"+session.FileID+"/complete", tokenB, nil)
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
}

func TestFile_InvalidULID(t *testing.T) {
	engine, mem, _ := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := getAuth(t, engine, "/api/v1/files/not-a-ulid", token)
	assertAPIError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}
