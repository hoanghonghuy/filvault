package main

import (
	"strings"
	"testing"
)

func TestReadCredentialsPipedInput(t *testing.T) {
	// Piped stdin carries both lines; a single shared reader must see both
	// (a second bufio.Reader would hit EOF after the first drained the pipe).
	var in strings.Builder
	in.WriteString("dev@filvault.com\nlocal-cli-test-password\n")

	email, password, err := readCredentials(strings.NewReader(in.String()))
	if err != nil {
		t.Fatalf("readCredentials: %v", err)
	}
	if email != "dev@filvault.com" {
		t.Fatalf("email=%q want dev@filvault.com", email)
	}
	if password != "local-cli-test-password" {
		t.Fatalf("password=%q want local-cli-test-password", password)
	}
}

func TestReadCredentialsNoEchoTrim(t *testing.T) {
	email, password, err := readCredentials(strings.NewReader("  a@b.co  \n pw1 \n"))
	if err != nil {
		t.Fatalf("readCredentials: %v", err)
	}
	if email != "a@b.co" || password != "pw1" {
		t.Fatalf("got %q / %q, want trimmed values", email, password)
	}
}
