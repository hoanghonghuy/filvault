package auth_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"filnest/internal/platform/mailer"
)

func TestProfile_PatchDisplayName(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := patchAuthJSON(t, engine, "/api/v1/users/me", token, map[string]any{
		"displayName": "New Name",
	})
	if code != http.StatusOK {
		t.Fatalf("patch status=%d body=%s", code, body)
	}
	if !contains(body, "New Name") {
		t.Fatalf("patch response=%s", body)
	}

	code, body = getAuth(t, engine, "/api/v1/users/me", token)
	if code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", code, body)
	}
	if !contains(body, `"displayName":"New Name"`) && !contains(body, `"displayName": "New Name"`) {
		t.Fatalf("me displayName not updated: %s", body)
	}
}

func TestProfile_PatchDisplayName_EmptyRejected(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := patchAuthJSON(t, engine, "/api/v1/users/me", token, map[string]any{
		"displayName": "   ",
	})
	assertError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func TestProfile_ChangePassword_RevokesAllRefresh(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	email := uniqueEmail()
	token := registerVerified(t, engine, mem, email)

	_, body := postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "password1",
	})
	other := decodeSession(t, body)

	code, body := postAuthJSON(t, engine, "/api/v1/users/me/password", token, map[string]any{
		"currentPassword": "password1",
		"newPassword":     "newpass99",
	})
	if code != http.StatusOK {
		t.Fatalf("change password status=%d body=%s", code, body)
	}
	var changed struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := decodeBody(body, &changed); err != nil {
		t.Fatalf("decode: %v body=%s", err, body)
	}
	if changed.AccessToken == "" || changed.RefreshToken == "" {
		t.Fatalf("missing tokens: %s", body)
	}

	code, body = postJSON(t, engine, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": other.RefreshToken,
	})
	assertError(t, code, body, http.StatusUnauthorized, "UNAUTHORIZED")

	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "password1",
	})
	assertError(t, code, body, http.StatusUnauthorized, "UNAUTHORIZED")

	code, body = postJSON(t, engine, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "newpass99",
	})
	if code != http.StatusOK {
		t.Fatalf("login new password status=%d body=%s", code, body)
	}
}

func TestProfile_ChangePassword_WrongCurrent(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuthJSON(t, engine, "/api/v1/users/me/password", token, map[string]any{
		"currentPassword": "wrong",
		"newPassword":     "newpass99",
	})
	assertError(t, code, body, http.StatusUnauthorized, "UNAUTHORIZED")
}

func TestProfile_ChangePassword_ShortNew(t *testing.T) {
	engine, mem := newAuthEngineWithMailer(t, "secret-invite")
	token := registerVerified(t, engine, mem, uniqueEmail())

	code, body := postAuthJSON(t, engine, "/api/v1/users/me/password", token, map[string]any{
		"currentPassword": "password1",
		"newPassword":     "short",
	})
	assertError(t, code, body, http.StatusBadRequest, "VALIDATION_ERROR")
}

func registerVerified(t *testing.T, engine http.Handler, mem *mailer.Memory, email string) string {
	t.Helper()
	_, body := postJSON(t, engine, "/api/v1/auth/register", map[string]any{
		"email":       email,
		"password":    "password1",
		"displayName": "User",
		"inviteCode":  "secret-invite",
	})
	code, _ := postJSON(t, engine, "/api/v1/auth/resend-verification", map[string]any{"email": email})
	if code != http.StatusNoContent {
		t.Fatalf("resend status=%d", code)
	}
	code, _ = postJSON(t, engine, "/api/v1/auth/verify-email", map[string]any{
		"email": email,
		"code":  mem.LastCode(email),
	})
	if code != http.StatusNoContent {
		t.Fatalf("verify status=%d", code)
	}
	session := decodeSession(t, body)
	return session.AccessToken
}

func patchAuthJSON(t *testing.T, engine http.Handler, path, token string, payload map[string]any) (int, string) {
	t.Helper()
	return doJSON(t, engine, http.MethodPatch, path, token, payload)
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func decodeBody(body string, v any) error {
	return json.Unmarshal([]byte(body), v)
}
