package auth_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"filvault/internal/app"
	"filvault/internal/platform/config"
	"filvault/internal/platform/mailer"
	"filvault/internal/platform/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestVerifyEmail_Flow(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	email := uniqueEmail()

	_, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       email,
		"password":    "password1",
		"displayName": "Huy",
		"inviteCode":  "secret-invite",
	})
	session := decodeSession(t, body)
	if session.User.EmailVerified {
		t.Fatal("new user must not be verified")
	}

	code, out := postJSON(t, engine, "/api/v1/auth/resend-verification", map[string]any{
		"email": email,
	})
	if code != http.StatusNoContent {
		t.Fatalf("resend status=%d body=%s", code, out)
	}
	got := mem.LastCode(email)
	if len(got) != 6 {
		t.Fatalf("expected 6-digit code, got %q", got)
	}

	code, out = postJSON(t, engine, "/api/v1/auth/verify-email", map[string]any{
		"email": email,
		"code":  "000000",
	})
	assertError(t, code, out, http.StatusBadRequest, "VALIDATION_ERROR")

	code, out = postJSON(t, engine, "/api/v1/auth/verify-email", map[string]any{
		"email": email,
		"code":  got,
	})
	if code != http.StatusNoContent {
		t.Fatalf("verify status=%d body=%s", code, out)
	}

	code, out = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "password1",
	})
	if code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", code, out)
	}
	login := decodeSession(t, out)
	if !login.User.EmailVerified {
		t.Fatal("user should be verified after verify-email")
	}

	code, out = getAuth(t, engine, "/api/v1/users/me", login.AccessToken)
	if code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", code, out)
	}
}

func TestResend_UnknownEmailStill204(t *testing.T) {
	engine, _ := newAuthEngineWithMailer(t, "secret-invite")
	code, out := postJSON(t, engine, "/api/v1/auth/resend-verification", map[string]any{
		"email": "nobody-" + uniqueEmail(),
	})
	if code != http.StatusNoContent {
		t.Fatalf("resend unknown status=%d body=%s", code, out)
	}
}

func TestVerify_UnknownEmailValidation(t *testing.T) {
	engine, _ := newAuthEngineWithMailer(t, "secret-invite")
	code, out := postJSON(t, engine, "/api/v1/auth/verify-email", map[string]any{
		"email": "nobody-" + uniqueEmail(),
		"code":  "123456",
	})
	assertError(t, code, out, http.StatusBadRequest, "VALIDATION_ERROR")
}

func newAuthEngineWithMailer(t *testing.T, invite string) (*gin.Engine, *mailer.Memory) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	url := envDBURL(t)
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

	mem := mailer.NewMemory()
	engine := app.NewWithMailer(config.Config{
		InviteCode:                invite,
		JWTSecret:                 "test-jwt-secret-not-for-prod",
		DefaultTrashAutoDelete:    false,
		DefaultTrashRetentionDays: config.DefaultTrashRetentionDays,
		MetadataStore:             "postgres",
	}, pool, mem)
	return engine, mem
}

func envDBURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("FILVAULT_DATABASE_URL")
	if url == "" {
		t.Fatal("FILVAULT_DATABASE_URL is required")
	}
	return url
}
