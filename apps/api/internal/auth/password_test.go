package auth

import (
	"strings"
	"testing"
)

func TestHashPassword_Argon2id(t *testing.T) {
	hash, err := hashPassword("password1")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected argon2id hash, got %q", hash)
	}
	if !verifyPassword(hash, "password1") {
		t.Fatal("verifyPassword should accept correct password")
	}
	if verifyPassword(hash, "wrong-password") {
		t.Fatal("verifyPassword should reject wrong password")
	}
}
