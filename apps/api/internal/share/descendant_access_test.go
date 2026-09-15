package share_test

import (
	"net/http"
	"testing"
)

func TestShares_DescendantFolderBrowseAndDownload(t *testing.T) {
	engine, mem, objs := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail("owner"))
	emailB := uniqueEmail("recipient")
	tokenB := registerVerified(t, engine, mem, emailB)

	root := createNestedFolder(t, engine, tokenA, "Shared root", "")
	child := createNestedFolder(t, engine, tokenA, "Child", root)
	grandchild := createNestedFolder(t, engine, tokenA, "Grandchild", child)
	fileID := uploadFile(t, engine, objs, tokenA, grandchild)

	code, body := postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   root,
		"email":        emailB,
	})
	if code != http.StatusCreated {
		t.Fatalf("share root status=%d body=%s", code, body)
	}

	for _, folderID := range []string{root, child, grandchild} {
		code, body = getAuth(t, engine, "/api/v1/shared/folders/"+folderID, tokenB)
		if code != http.StatusOK {
			t.Fatalf("browse descendant %s status=%d body=%s", folderID, code, body)
		}
	}

	code, body = getAuth(t, engine, "/api/v1/shared/files/"+fileID+"/download", tokenB)
	if code != http.StatusOK {
		t.Fatalf("download descendant status=%d body=%s", code, body)
	}
	var download struct {
		DownloadURL string `json:"downloadUrl"`
	}
	decodeJSON(t, body, &download)
	if download.DownloadURL == "" {
		t.Fatalf("download descendant url empty: %s", body)
	}
}

func TestShares_DescendantAccessDoesNotCrossUnsharedTree(t *testing.T) {
	engine, mem, objs := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail("owner"))
	emailB := uniqueEmail("recipient")
	tokenB := registerVerified(t, engine, mem, emailB)

	sharedRoot := createNestedFolder(t, engine, tokenA, "Shared root", "")
	unsharedRoot := createNestedFolder(t, engine, tokenA, "Private root", "")
	privateChild := createNestedFolder(t, engine, tokenA, "Private child", unsharedRoot)
	privateFileID := uploadFile(t, engine, objs, tokenA, privateChild)

	code, body := postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   sharedRoot,
		"email":        emailB,
	})
	if code != http.StatusCreated {
		t.Fatalf("share root status=%d body=%s", code, body)
	}

	code, body = getAuth(t, engine, "/api/v1/shared/folders/"+privateChild, tokenB)
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")

	code, body = getAuth(t, engine, "/api/v1/shared/files/"+privateFileID+"/download", tokenB)
	assertAPIError(t, code, body, http.StatusNotFound, "NOT_FOUND")
}

func createNestedFolder(t *testing.T, engine http.Handler, token, name, parentID string) string {
	t.Helper()
	payload := map[string]any{"name": name}
	if parentID != "" {
		payload["parentId"] = parentID
	}
	code, body := postAuth(t, engine, "/api/v1/folders", token, payload)
	if code != http.StatusCreated {
		t.Fatalf("create folder %q status=%d body=%s", name, code, body)
	}
	var folder folderResp
	decodeJSON(t, body, &folder)
	if folder.ID == "" {
		t.Fatalf("create folder %q returned empty id: %s", name, body)
	}
	return folder.ID
}
