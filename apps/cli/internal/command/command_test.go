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

func TestLogout(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/logout" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	cleared := false
	out, _ := run(t, srv, func(c *Commands) error {
		return c.Logout("refresh-1", func() error { cleared = true; return nil })
	})
	if !strings.Contains(out, "Logged out") {
		t.Fatalf("logout output missing: %q", out)
	}
	if !cleared {
		t.Fatal("clear callback not called")
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

func TestMkdir(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/folders" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Docs" {
			t.Fatalf("unexpected name %v", body["name"])
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": "f1", "name": "Docs"})
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Mkdir("Docs", "") })
	if !strings.Contains(out, "Docs") {
		t.Fatalf("mkdir output missing name: %q", out)
	}
}

func TestRmFile(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/browser":
			json.NewEncoder(w).Encode(map[string]any{
				"folders": []any{},
				"files":   []map[string]any{{"id": "f1", "name": "a.pdf", "sizeBytes": 5}},
			})
		case r.URL.Path == "/files/f1" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Rm("a.pdf") })
	if !strings.Contains(out, "a.pdf") {
		t.Fatalf("rm output missing name: %q", out)
	}
}

func TestRmFolder(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/browser":
			json.NewEncoder(w).Encode(map[string]any{
				"folders": []map[string]any{{"id": "f1", "name": "Docs"}},
				"files":   []any{},
			})
		case r.URL.Path == "/folders/f1" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Rm("Docs") })
	if !strings.Contains(out, "Docs") {
		t.Fatalf("rm output missing name: %q", out)
	}
}

func TestMvRename(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/browser":
			json.NewEncoder(w).Encode(map[string]any{
				"folders": []any{},
				"files":   []map[string]any{{"id": "f1", "name": "a.pdf", "sizeBytes": 5}},
			})
		case r.URL.Path == "/files/f1" && r.Method == http.MethodPatch:
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["name"] != "b.pdf" {
				t.Fatalf("unexpected name %v", body["name"])
			}
			json.NewEncoder(w).Encode(map[string]any{"id": "f1", "name": "b.pdf"})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Mv("a.pdf", "b.pdf", "") })
	if !strings.Contains(out, "a.pdf") {
		t.Fatalf("mv output missing name: %q", out)
	}
}

func TestMvMove(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/browser":
			json.NewEncoder(w).Encode(map[string]any{
				"folders": []any{},
				"files":   []map[string]any{{"id": "f1", "name": "a.pdf", "sizeBytes": 5}},
			})
		case r.URL.Path == "/files/f1" && r.Method == http.MethodPatch:
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["folderId"] != "dest" {
				t.Fatalf("unexpected folderId %v", body["folderId"])
			}
			json.NewEncoder(w).Encode(map[string]any{"id": "f1", "name": "a.pdf"})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Mv("a.pdf", "", "dest") })
	if !strings.Contains(out, "a.pdf") {
		t.Fatalf("mv output missing name: %q", out)
	}
}

func TestTrash(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trash" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"folders": []map[string]any{{"id": "f1", "name": "Docs", "type": "folder"}},
			"files":   []map[string]any{{"id": "x1", "name": "a.pdf", "type": "file", "sizeBytes": 5}},
		})
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Trash() })
	if !strings.Contains(out, "Docs") || !strings.Contains(out, "a.pdf") {
		t.Fatalf("trash output missing entries: %q", out)
	}
}

func TestRestore(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/trash":
			json.NewEncoder(w).Encode(map[string]any{
				"folders": []any{},
				"files":   []map[string]any{{"id": "x1", "name": "a.pdf", "type": "file", "sizeBytes": 5}},
			})
		case r.URL.Path == "/files/x1/restore" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Restore("a.pdf") })
	if !strings.Contains(out, "a.pdf") {
		t.Fatalf("restore output missing name: %q", out)
	}
}

func TestSearch(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "doc" {
			t.Fatalf("unexpected q %q", r.URL.Query().Get("q"))
		}
		json.NewEncoder(w).Encode(map[string]any{
			"folders": []map[string]any{{"id": "f1", "name": "Docs"}},
			"files":   []map[string]any{{"id": "x1", "name": "doc.pdf", "sizeBytes": 5}},
		})
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Search("doc") })
	if !strings.Contains(out, "Docs") || !strings.Contains(out, "doc.pdf") {
		t.Fatalf("search output missing entries: %q", out)
	}
}

func TestStorage(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storage" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"usedBytes": 1024, "quotaBytes": 1048576})
	})

	out, _ := run(t, srv, func(c *Commands) error { return c.Storage() })
	if !strings.Contains(out, "1.0 KiB") || !strings.Contains(out, "1.0 MiB") {
		t.Fatalf("storage output missing sizes: %q", out)
	}
}

func TestCd(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/browser" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"folder":     nil,
			"breadcrumb": []any{},
			"folders":    []map[string]any{{"id": "f1", "name": "Docs"}},
			"files":      []any{},
		})
	})

	var saved string
	var out bytes.Buffer
	c := client.New(srv.URL)
	cmds := &Commands{Client: c, Out: &out, Err: &out, SaveFolder: func(id string) error { saved = id; return nil }}
	if err := cmds.Cd("Docs"); err != nil {
		t.Fatalf("cd error: %v", err)
	}
	if saved != "f1" {
		t.Fatalf("saved folder = %q, want f1", saved)
	}
	if !strings.Contains(out.String(), "Docs") {
		t.Fatalf("cd output missing name: %q", out.String())
	}
}

func TestCdUp(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/browser" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("folderId") != "f1" {
			t.Fatalf("unexpected folderId %q", r.URL.Query().Get("folderId"))
		}
		json.NewEncoder(w).Encode(map[string]any{
			"folder":     map[string]any{"id": "f1", "name": "Docs", "parentId": "root1"},
			"breadcrumb": []any{},
			"folders":    []any{},
			"files":      []any{},
		})
	})

	var saved string
	var out bytes.Buffer
	c := client.New(srv.URL)
	cmds := &Commands{Client: c, Out: &out, Err: &out, CurrentFolderID: "f1", SaveFolder: func(id string) error { saved = id; return nil }}
	if err := cmds.Cd(".."); err != nil {
		t.Fatalf("cd .. error: %v", err)
	}
	if saved != "root1" {
		t.Fatalf("saved folder = %q, want root1", saved)
	}
}

func TestCdRoot(t *testing.T) {
	var saved string
	var out bytes.Buffer
	cmds := &Commands{Out: &out, Err: &out, CurrentFolderID: "f1", SaveFolder: func(id string) error { saved = id; return nil }}
	if err := cmds.Cd(""); err != nil {
		t.Fatalf("cd root error: %v", err)
	}
	if saved != "" {
		t.Fatalf("saved folder = %q, want empty", saved)
	}
}

func TestPwd(t *testing.T) {
	var out bytes.Buffer
	cmds := &Commands{Out: &out, Err: &out, CurrentFolderID: "f1"}
	if err := cmds.Pwd(); err != nil {
		t.Fatalf("pwd error: %v", err)
	}
	if !strings.Contains(out.String(), "f1") {
		t.Fatalf("pwd output missing id: %q", out.String())
	}
}
