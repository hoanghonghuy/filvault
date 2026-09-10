package auth_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPasswordRecovery_FlowAndSessionRevocation(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	email := uniqueEmail()

	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       email,
		"password":    "password1",
		"displayName": "Recovery User",
		"inviteCode":  "secret-invite",
	})
	if code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", code, body)
	}
	original := decodeSession(t, body)

	code, body = postJSON(t, engine, "/api/v1/auth/password/forgot", map[string]any{"email": email})
	if code != http.StatusNoContent {
		t.Fatalf("forgot status=%d body=%s", code, body)
	}
	resetToken := mem.LastCode(email)
	if len(resetToken) < 40 {
		t.Fatalf("expected high-entropy reset token, got length %d", len(resetToken))
	}

	code, body = postJSON(t, engine, "/api/v1/auth/password/reset", map[string]any{
		"token":       resetToken,
		"newPassword": "new-password-2",
	})
	if code != http.StatusNoContent {
		t.Fatalf("reset status=%d body=%s", code, body)
	}

	code, body = postJSON(t, engine, "/api/v1/auth/password/reset", map[string]any{
		"token":       resetToken,
		"newPassword": "another-password-3",
	})
	assertError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email": email, "password": "password1",
	})
	assertError(t, code, body, http.StatusUnauthorized, "UNAUTHORIZED")

	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email": email, "password": "new-password-2",
	})
	if code != http.StatusOK {
		t.Fatalf("login with reset password status=%d body=%s", code, body)
	}

	code, body = postJSON(t, engine, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": original.RefreshToken,
	})
	assertError(t, code, body, http.StatusUnauthorized, "UNAUTHORIZED")
}

func TestPasswordRecovery_UnknownEmailIsIndistinguishable(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	knownEmail := uniqueEmail()

	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       knownEmail,
		"password":    "password1",
		"displayName": "Known User",
		"inviteCode":  "secret-invite",
	})
	if code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", code, body)
	}

	code, knownBody := postJSON(t, engine, "/api/v1/auth/password/forgot", map[string]any{"email": knownEmail})
	if code != http.StatusNoContent || knownBody != "" {
		t.Fatalf("known forgot status=%d body=%q", code, knownBody)
	}

	unknownEmail := "nobody-" + uniqueEmail()
	code, unknownBody := postJSON(t, engine, "/api/v1/auth/password/forgot", map[string]any{"email": unknownEmail})
	if code != http.StatusNoContent || unknownBody != knownBody {
		t.Fatalf("unknown forgot status=%d body=%q; known body=%q", code, unknownBody, knownBody)
	}
	if mem.LastCode(unknownEmail) != "" {
		t.Fatal("unknown account must not receive reset mail")
	}
}

func TestPasswordRecovery_NewRequestInvalidatesPreviousToken(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	email := uniqueEmail()
	registerRecoveryUser(t, engine, email)

	postForgot(t, engine, email)
	first := mem.LastCode(email)
	postForgot(t, engine, email)
	second := mem.LastCode(email)
	if first == second || first == "" || second == "" {
		t.Fatal("expected each forgot request to replace the token")
	}

	code, body := postJSON(t, engine, "/api/v1/auth/password/reset", map[string]any{
		"token": first, "newPassword": "new-password-2",
	})
	assertError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")

	code, body = postJSON(t, engine, "/api/v1/auth/password/reset", map[string]any{
		"token": second, "newPassword": "new-password-2",
	})
	if code != http.StatusNoContent {
		t.Fatalf("latest reset token status=%d body=%s", code, body)
	}
}

func TestPasswordRecovery_ExpiredTokenIsRejected(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	email := uniqueEmail()
	registerRecoveryUser(t, engine, email)
	postForgot(t, engine, email)
	token := mem.LastCode(email)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, envDBURL(t))
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, `UPDATE users SET password_reset_expires_at = now() - interval '1 minute' WHERE email = $1`, email); err != nil {
		t.Fatalf("expire reset token: %v", err)
	}

	code, body := postJSON(t, engine, "/api/v1/auth/password/reset", map[string]any{
		"token": token, "newPassword": "new-password-2",
	})
	assertError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func registerRecoveryUser(t *testing.T, engine http.Handler, email string) {
	t.Helper()
	code, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       email,
		"password":    "password1",
		"displayName": "Recovery User",
		"inviteCode":  "secret-invite",
	})
	if code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", code, body)
	}
}

func postForgot(t *testing.T, engine http.Handler, email string) {
	t.Helper()
	code, body := postJSON(t, engine, "/api/v1/auth/password/forgot", map[string]any{"email": email})
	if code != http.StatusNoContent {
		t.Fatalf("forgot status=%d body=%s", code, body)
	}
}
