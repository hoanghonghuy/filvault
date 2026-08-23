package share_test

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestShares_UnknownEmailSendsInviteNot404 covers the anti-probing contract:
// sharing to an unregistered email must succeed with 201 and send an invite
// mail, never reveal whether the account exists.
func TestShares_UnknownEmailSendsInviteNot404(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail("a"))

	folderID := createFolder(t, engine, tokenA, "Invited")
	code, body := postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   folderID,
		"email":        "stranger@example.com",
	})
	if code != http.StatusCreated {
		t.Fatalf("share status=%d want 201 body=%s", code, body)
	}
	var created struct {
		ID      string `json:"id"`
		Invited bool   `json:"invited"`
	}
	decodeJSON(t, body, &created)
	if !created.Invited {
		t.Fatalf("expected invited=true, got %s", body)
	}
	if created.ID != "" {
		t.Fatalf("unknown email must not create a share row, got id=%q", created.ID)
	}
	if !strings.Contains(mem.LastCode("stranger@example.com"), "Filvault") {
		t.Fatalf("invite mail not sent: %q", mem.LastCode("stranger@example.com"))
	}
}

// TestShares_CreateResponseIncludesRecipient locks the recipient payload the
// web UI needs after a successful share.
func TestShares_CreateResponseIncludesRecipient(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail("a"))
	emailB := uniqueEmail("b")
	registerVerified(t, engine, mem, emailB)

	folderID := createFolder(t, engine, tokenA, "Docs")

	code, body := postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   folderID,
		"email":        emailB,
	})
	if code != http.StatusCreated {
		t.Fatalf("share status=%d body=%s", code, body)
	}
	var created struct {
		ID           string `json:"id"`
		ResourceType string `json:"resourceType"`
		ResourceID   string `json:"resourceId"`
		Recipient    struct {
			ID          string `json:"id"`
			Email       string `json:"email"`
			DisplayName string `json:"displayName"`
		} `json:"recipient"`
		CreatedAt string `json:"createdAt"`
	}
	decodeJSON(t, body, &created)
	if created.Recipient.Email != emailB || created.Recipient.ID == "" {
		t.Fatalf("recipient missing: %s", body)
	}
	if _, err := time.Parse(time.RFC3339Nano, created.CreatedAt); err != nil {
		t.Fatalf("createdAt format: %v", err)
	}
}

// TestShares_ActivityRecordsInternalShare verifies share.created / share.revoked
// land in the owner's activity feed for internal shares too.
func TestShares_ActivityRecordsInternalShare(t *testing.T) {
	engine, mem, _ := newEngine(t)
	tokenA := registerVerified(t, engine, mem, uniqueEmail("a"))
	emailB := uniqueEmail("b")
	registerVerified(t, engine, mem, emailB)

	folderID := createFolder(t, engine, tokenA, "Logged")

	code, body := postAuth(t, engine, "/api/v1/shares", tokenA, map[string]any{
		"resourceType": "folder",
		"resourceId":   folderID,
		"email":        emailB,
	})
	if code != http.StatusCreated {
		t.Fatalf("share status=%d body=%s", code, body)
	}
	var created shareResp
	decodeJSON(t, body, &created)

	page := activityPageOf(t, engine, tokenA)
	found := false
	for _, ev := range page {
		if ev.Type == "share.created" && ev.TargetName == "Logged" {
			found = true
		}
	}
	if !found {
		t.Fatalf("share.created event missing")
	}

	// Revoke also records.
	code, _ = deleteAuth(t, engine, "/api/v1/shares/"+created.ID, tokenA)
	if code != http.StatusNoContent {
		t.Fatalf("revoke status=%d", code)
	}
	page = activityPageOf(t, engine, tokenA)
	revoked := false
	for _, ev := range page {
		if ev.Type == "share.revoked" && ev.TargetName == "Logged" {
			revoked = true
		}
	}
	if !revoked {
		t.Fatalf("share.revoked event missing")
	}
}

type simpleEvent struct {
	Type       string `json:"type"`
	TargetName string `json:"targetName"`
}

func activityPageOf(t *testing.T, engine http.Handler, token string) []simpleEvent {
	t.Helper()
	code, body := getAuth(t, engine, "/api/v1/activity?limit=50", token)
	if code != http.StatusOK {
		t.Fatalf("activity status=%d body=%s", code, body)
	}
	var page struct {
		Events []simpleEvent `json:"events"`
	}
	decodeJSON(t, body, &page)
	return page.Events
}

func createFolder(t *testing.T, engine http.Handler, token, name string) string {
	t.Helper()
	_, body := postAuth(t, engine, "/api/v1/folders", token, map[string]any{"name": name})
	var folder folderResp
	decodeJSON(t, body, &folder)
	return folder.ID
}
