package auth

import "testing"

func TestPasswordResetTokenShape(t *testing.T) {
	token, err := newPasswordResetToken()
	if err != nil {
		t.Fatalf("newPasswordResetToken: %v", err)
	}
	if !validPasswordResetToken(token) {
		t.Fatal("generated password reset token must be accepted")
	}
	for _, invalid := range []string{"", "short", token + "=", token[:len(token)-1], "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"} {
		if validPasswordResetToken(invalid) {
			t.Fatalf("malformed token accepted: %q", invalid)
		}
	}
}
