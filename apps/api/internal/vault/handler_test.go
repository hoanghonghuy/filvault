package vault_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"filvault/internal/auth"
	"filvault/internal/vault"

	"github.com/gin-gonic/gin"
)

func setupTestRouter(svc *vault.Service, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := vault.NewHandler(svc)

	fakeAuth := func(c *gin.Context) {
		c.Set(auth.CtxUserID, userID)
		c.Next()
	}

	v1 := r.Group("/api/v1")
	handler.RegisterRoutes(v1, fakeAuth)
	return r
}

func TestVaultHandler_Endpoints(t *testing.T) {
	repo := &mockRepo{}
	verifier := &mockVerifier{validPassword: "AccountPass123"}
	tokens := vault.NewTokens("jwt-secret-key-1234567890123456", 1*time.Hour)
	svc := vault.NewService(repo, verifier, tokens)

	userID := "01M1X1Q2YTJ4RBHWK82VE2YCCN"
	r := setupTestRouter(svc, userID)

	// 1. GET /api/v1/vault/status -> initialized: false
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vault/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var st vault.Status
	_ = json.Unmarshal(w.Body.Bytes(), &st)
	if st.Initialized || st.Unlocked {
		t.Fatalf("expected not initialized, got %+v", st)
	}

	// 2. POST /api/v1/vault/setup -> returns token
	body, _ := json.Marshal(map[string]string{"pin": "123456"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/vault/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("setup failed with code %d: %s", w.Code, w.Body.String())
	}
	var session vault.Session
	_ = json.Unmarshal(w.Body.Bytes(), &session)
	if !session.Unlocked || session.Token == "" {
		t.Fatalf("expected unlocked session, got %+v", session)
	}
	vaultToken := session.Token

	// 3. GET /api/v1/vault/files without token -> 401
	req = httptest.NewRequest(http.MethodGet, "/api/v1/vault/files", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", w.Code)
	}

	// 4. GET /api/v1/vault/files with token -> 200
	req = httptest.NewRequest(http.MethodGet, "/api/v1/vault/files", nil)
	req.Header.Set("X-Vault-Token", vaultToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with token, got %d", w.Code)
	}

	// 5. POST /api/v1/vault/items -> move file into vault
	moveBody, _ := json.Marshal(map[string]any{"fileIds": []string{"file-abc"}})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/vault/items", bytes.NewReader(moveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", vaultToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on move items, got %d: %s", w.Code, w.Body.String())
	}

	// 6. POST /api/v1/vault/unlock with wrong PIN -> 401
	unlockBody, _ := json.Marshal(map[string]string{"pin": "wrong"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/vault/unlock", bytes.NewReader(unlockBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on wrong PIN, got %d", w.Code)
	}
}
