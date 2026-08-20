package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoParsesErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"code": "CONFLICT", "message": "name exists"},
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	err := c.Do(http.MethodGet, "/x", nil, nil)
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T: %v", err, err)
	}
	if apiErr.Code != "CONFLICT" || apiErr.Message != "name exists" || apiErr.Status != 409 {
		t.Fatalf("unexpected error: %+v", apiErr)
	}
}

func TestDoRefreshesOn401Once(t *testing.T) {
	refreshCalls := 0
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			refreshCalls++
			json.NewEncoder(w).Encode(map[string]string{
				"accessToken":  "new-access",
				"refreshToken": "new-refresh",
			})
		case "/me":
			if r.Header.Get("Authorization") == "Bearer new-access" {
				json.NewEncoder(w).Encode(map[string]string{"id": "u1"})
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{"code": "UNAUTHORIZED", "message": "expired"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := New(srv.URL)
	c.Access = "old-access"
	c.Refresh = "old-refresh"

	var out map[string]string
	if err := c.Do(http.MethodGet, "/me", nil, &out); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if out["id"] != "u1" {
		t.Fatalf("unexpected body: %+v", out)
	}
	if refreshCalls != 1 {
		t.Fatalf("refresh called %d times, want 1", refreshCalls)
	}
	if c.Access != "new-access" || c.Refresh != "new-refresh" {
		t.Fatalf("tokens not updated: %+v", c)
	}
}

func TestDoNoContentReturnsNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL)
	if err := c.Do(http.MethodPost, "/logout", map[string]string{"refreshToken": "r"}, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
}

func TestRefreshFailureCallsOnAuthFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"code": "UNAUTHORIZED", "message": "bad refresh"},
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	c.Access = "a"
	c.Refresh = "r"
	authFailed := false
	c.OnAuthFail = func() { authFailed = true }

	err := c.Do(http.MethodGet, "/me", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !authFailed {
		t.Fatal("OnAuthFail not called")
	}
}
