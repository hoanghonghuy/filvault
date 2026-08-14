package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"filnest/internal/app"
	"filnest/internal/platform/config"
	"filnest/internal/platform/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestRegister_MissingInvite(t *testing.T) {
	engine, _ := newAuthEngine(t, "secret-invite")
	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       uniqueEmail(),
		"password":    "password1",
		"displayName": "Huy",
	})
	assertError(t, code, body, http.StatusForbidden, "FORBIDDEN")
}

func TestRegister_WrongInvite(t *testing.T) {
	engine, _ := newAuthEngine(t, "secret-invite")
	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       uniqueEmail(),
		"password":    "password1",
		"displayName": "Huy",
		"inviteCode":  "nope",
	})
	assertError(t, code, body, http.StatusForbidden, "FORBIDDEN")
}

func TestRegister_DisabledWhenInviteEmpty(t *testing.T) {
	engine, _ := newAuthEngine(t, "")
	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       uniqueEmail(),
		"password":    "password1",
		"displayName": "Huy",
		"inviteCode":  "anything",
	})
	assertError(t, code, body, http.StatusForbidden, "REGISTER_DISABLED")
}

func TestRegister_ShortPassword(t *testing.T) {
	engine, _ := newAuthEngine(t, "secret-invite")
	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       uniqueEmail(),
		"password":    "short",
		"displayName": "Huy",
		"inviteCode":  "secret-invite",
	})
	assertError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func TestRegister_Login_Me_Refresh_Logout(t *testing.T) {
	engine, _ := newAuthEngine(t, "secret-invite")
	email := uniqueEmail()

	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       strings.ToUpper(email),
		"password":    "password1",
		"displayName": "Huy",
		"inviteCode":  "secret-invite",
	})
	if code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", code, body)
	}
	session := decodeSession(t, body)
	if session.User.Email != email {
		t.Fatalf("email stored as %q, want lowercase %q", session.User.Email, email)
	}
	if session.User.EmailVerified {
		t.Fatal("new user must not be verified")
	}
	if session.User.StorageQuota != config.DefaultStorageQuotaBytes {
		t.Fatalf("quota=%d", session.User.StorageQuota)
	}
	if strings.Contains(body, "passwordHash") || strings.Contains(body, "password_hash") {
		t.Fatal("password hash leaked")
	}
	if session.AccessToken == "" || session.RefreshToken == "" {
		t.Fatal("missing tokens")
	}

	code, body = postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       email,
		"password":    "password1",
		"displayName": "Huy",
		"inviteCode":  "secret-invite",
	})
	assertError(t, code, body, http.StatusConflict, "CONFLICT")

	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "wrong-password",
	})
	assertError(t, code, body, http.StatusUnauthorized, "UNAUTHORIZED")

	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    "missing-" + email,
		"password": "password1",
	})
	assertError(t, code, body, http.StatusUnauthorized, "UNAUTHORIZED")

	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "password1",
	})
	if code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", code, body)
	}
	login := decodeSession(t, body)

	code, body = getAuth(t, engine, "/api/v1/users/me", "")
	assertError(t, code, body, http.StatusUnauthorized, "UNAUTHORIZED")

	code, body = getAuth(t, engine, "/api/v1/users/me", login.AccessToken)
	if code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", code, body)
	}
	var me struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		DisplayName   string `json:"displayName"`
		EmailVerified bool   `json:"emailVerified"`
	}
	if err := json.Unmarshal([]byte(body), &me); err != nil {
		t.Fatalf("me json: %v", err)
	}
	if me.ID != login.User.ID || me.Email != email || me.DisplayName != "Huy" || me.EmailVerified {
		t.Fatalf("me payload=%s", body)
	}

	code, body = postJSON(t, engine, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": login.RefreshToken,
	})
	if code != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", code, body)
	}
	refreshed := decodeSession(t, body)
	if refreshed.AccessToken == "" {
		t.Fatal("refresh missing access token")
	}

	code, body = postAuthJSON(t, engine, "/api/v1/auth/logout", login.AccessToken, map[string]any{
		"refreshToken": login.RefreshToken,
	})
	if code != http.StatusNoContent {
		t.Fatalf("logout status=%d body=%s", code, body)
	}

	code, body = postJSON(t, engine, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": login.RefreshToken,
	})
	assertError(t, code, body, http.StatusUnauthorized, "UNAUTHORIZED")
}

func TestLogout_RevokesOnlyOneSession(t *testing.T) {
	engine, _ := newAuthEngine(t, "secret-invite")
	email := uniqueEmail()
	_, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       email,
		"password":    "password1",
		"displayName": "Huy",
		"inviteCode":  "secret-invite",
	})
	a := decodeSession(t, body)
	_, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "password1",
	})
	b := decodeSession(t, body)

	code, out := postAuthJSON(t, engine, "/api/v1/auth/logout", a.AccessToken, map[string]any{
		"refreshToken": a.RefreshToken,
	})
	if code != http.StatusNoContent {
		t.Fatalf("logout status=%d body=%s", code, out)
	}

	code, out = postJSON(t, engine, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": a.RefreshToken,
	})
	assertError(t, code, out, http.StatusUnauthorized, "UNAUTHORIZED")

	code, out = postJSON(t, engine, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": b.RefreshToken,
	})
	if code != http.StatusOK {
		t.Fatalf("other session refresh status=%d body=%s", code, out)
	}
}

type sessionBody struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"emailVerified"`
		StorageQuota  int64  `json:"storageQuota"`
	} `json:"user"`
}

func decodeSession(t *testing.T, body string) sessionBody {
	t.Helper()
	var s sessionBody
	if err := json.Unmarshal([]byte(body), &s); err != nil {
		t.Fatalf("session json: %v body=%s", err, body)
	}
	return s
}

func assertError(t *testing.T, status int, body string, wantStatus int, wantCode string) {
	t.Helper()
	if status != wantStatus {
		t.Fatalf("status=%d want %d body=%s", status, wantStatus, body)
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("error json: %v body=%s", err, body)
	}
	if payload.Error.Code != wantCode {
		t.Fatalf("code=%q want %q body=%s", payload.Error.Code, wantCode, body)
	}
}

func newAuthEngine(t *testing.T, invite string) (*gin.Engine, *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	url := os.Getenv("FILNEST_DATABASE_URL")
	if url == "" {
		t.Fatal("FILNEST_DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	engine := app.New(config.Config{
		InviteCode:                invite,
		JWTSecret:                 "test-jwt-secret-not-for-prod",
		DefaultTrashAutoDelete:    false,
		DefaultTrashRetentionDays: config.DefaultTrashRetentionDays,
		MetadataStore:             "postgres",
	}, pool)
	return engine, pool
}

func uniqueEmail() string {
	return fmt.Sprintf("u-%d@example.com", time.Now().UnixNano())
}

func postJSON(t *testing.T, engine http.Handler, path string, payload map[string]any) (int, string) {
	t.Helper()
	return doJSON(t, engine, http.MethodPost, path, "", payload)
}

func postAuthJSON(t *testing.T, engine http.Handler, path, token string, payload map[string]any) (int, string) {
	t.Helper()
	return doJSON(t, engine, http.MethodPost, path, token, payload)
}

func getAuth(t *testing.T, engine http.Handler, path, token string) (int, string) {
	t.Helper()
	return doJSON(t, engine, http.MethodGet, path, token, nil)
}

func doJSON(t *testing.T, engine http.Handler, method, path, token string, payload map[string]any) (int, string) {
	t.Helper()
	var buf bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&buf).Encode(payload); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}
