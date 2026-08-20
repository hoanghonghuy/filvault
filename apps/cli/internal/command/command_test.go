package command

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filvault/cli/internal/client"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func run(t *testing.T, srv *httptest.Server, fn func(*Commands) error) (string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	c := client.New(srv.URL)
	cmds := &Commands{Client: c, Out: &out, Err: &errOut}
	if err := fn(cmds); err != nil {
		t.Fatalf("command error: %v", err)
	}
	return out.String(), errOut.String()
}

func TestLogin(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/login" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"user":         map[string]any{"id": "u1", "email": "dev@filvault.com"},
			"accessToken":  "access-1",
			"refreshToken": "refresh-1",
		})
	})

	var savedAccess, savedRefresh string
	out, _ := run(t, srv, func(c *Commands) error {
		return c.Login("dev@filvault.com", "pw", func(a, r string) error {
			savedAccess, savedRefresh = a, r
			return nil
		})
	})
	if !strings.Contains(out, "dev@filvault.com") {
		t.Fatalf("login output missing email: %q", out)
	}
	if savedAccess != "access-1" || savedRefresh != "refresh-1" {
		t.Fatalf("tokens not saved: %q %q", savedAccess, savedRefresh)
	}
}

func TestWhoami(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/me" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id": "u1", "email": "dev@filvault.com", "displayName": "Dev",
			"storageUsed": 1024, "storageQuota": 1048576,
		})
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Whoami() })
	if !strings.Contains(out, "dev@filvault.com") {
		t.Fatalf("whoami output missing email: %q", out)
	}
	if !strings.Contains(out, "1.0 KiB") {
		t.Fatalf("whoami output missing formatted size: %q", out)
	}
}

func TestLs(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/browser" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"folders": []map[string]any{{"id": "f1", "name": "Docs"}},
			"files":   []map[string]any{{"id": "x1", "name": "a.pdf", "sizeBytes": 2048}},
		})
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Ls("") })
	if !strings.Contains(out, "Docs") || !strings.Contains(out, "a.pdf") {
		t.Fatalf("ls output missing entries: %q", out)
	}
}

func TestUpload(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "report.pdf")
	if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	var putCalled bool
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/files/upload-sessions" && r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]string{
				"fileId": "f1", "uploadUrl": srv.URL + "/put", "expiresAt": "x",
			})
		case r.URL.Path == "/put" && r.Method == http.MethodPut:
			putCalled = true
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/files/f1/complete" && r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]string{"id": "f1", "name": "report.pdf"})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	out, _ := run(t, srv, func(c *Commands) error { return c.Upload(filePath, "") })
	if !strings.Contains(out, "report.pdf") {
		t.Fatalf("upload output missing name: %q", out)
	}
	if !putCalled {
		t.Fatal("presigned PUT not called")
	}
}

func TestDownload(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/browser":
			json.NewEncoder(w).Encode(map[string]any{
				"folders": []any{},
				"files":   []map[string]any{{"id": "f1", "name": "a.pdf", "sizeBytes": 5}},
			})
		case r.URL.Path == "/files/f1/download":
			json.NewEncoder(w).Encode(map[string]string{
				"downloadUrl": srv.URL + "/get", "expiresAt": "x",
			})
		case r.URL.Path == "/get":
			w.Write([]byte("hello"))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)

	out, _ := run(t, srv, func(c *Commands) error { return c.Download("a.pdf") })
	if !strings.Contains(out, "a.pdf") {
		t.Fatalf("download output missing name: %q", out)
	}
	data, err := os.ReadFile(filepath.Join(dir, "a.pdf"))
	if err != nil || string(data) != "hello" {
		t.Fatalf("downloaded content mismatch: %q err=%v", data, err)
	}
}
