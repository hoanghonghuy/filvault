package config

import (
	"strings"
	"testing"
)

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func baseProdEnv() map[string]string {
	return map[string]string{
		"FILVAULT_ENV":                   "production",
		"FILVAULT_DATABASE_URL":          "postgres://app:strong-db-pass@db.example.com:5432/filvault?sslmode=require",
		"FILVAULT_JWT_SECRET":            strings.Repeat("a", 32),
		"FILVAULT_INVITE_CODE":           "prod-invite-code-16",
		"FILVAULT_OBJECT_STORE":          "s3",
		"FILVAULT_S3_ENDPOINT":           "https://s3.amazonaws.com",
		"FILVAULT_S3_PUBLIC_ENDPOINT":    "https://cdn.example.com",
		"FILVAULT_S3_ACCESS_KEY":         "prod-s3-access-key-example-only",         // gitleaks:allow
		"FILVAULT_S3_SECRET_KEY":         "example-s3-secret-key-not-in-repo",       // gitleaks:allow
		"FILVAULT_CORS_ALLOWED_ORIGINS":  "https://app.example.com",
		"FILVAULT_MAILER":                "smtp",
		"FILVAULT_SMTP_HOST":             "smtp.example.com",
		"FILVAULT_SMTP_FROM":             "noreply@example.com",
		"FILVAULT_LIVEKIT_PUBLIC_URL":    "wss://livekit.example.com",
		"FILVAULT_LIVEKIT_API_KEY":       "prod-livekit-key",
		"FILVAULT_LIVEKIT_API_SECRET":    "prod-livekit-secret-value",
		"FILVAULT_SEED_DEV_USER":         "false",
	}
}

func TestLoad_DevelopmentAllowsPlaceholderSecrets(t *testing.T) {
	setEnv(t, map[string]string{
		"FILVAULT_ENV":          "development",
		"FILVAULT_DATABASE_URL": "postgres://filvault:replace-with-local-postgres-password@127.0.0.1:5435/filvault?sslmode=disable",
		"FILVAULT_JWT_SECRET":   "replace-with-local-jwt-secret-min-32-chars",
		"FILVAULT_OBJECT_STORE": "s3",
		"FILVAULT_S3_ENDPOINT":  "http://127.0.0.1:9002",
		"FILVAULT_S3_ACCESS_KEY":  "filvault",
		"FILVAULT_S3_SECRET_KEY":  "replace-with-local-s3-secret-key",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != EnvDevelopment {
		t.Fatalf("Env = %q, want development", cfg.Env)
	}
}

func TestLoad_ProductionRejectsPlaceholderJWT(t *testing.T) {
	env := baseProdEnv()
	env["FILVAULT_JWT_SECRET"] = "replace-with-local-jwt-secret-min-32-chars"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for placeholder JWT in production")
	}
	if !strings.Contains(err.Error(), "FILVAULT_JWT_SECRET") {
		t.Fatalf("error = %v, want JWT guidance", err)
	}
}

func TestLoad_ProductionRejectsShortJWT(t *testing.T) {
	env := baseProdEnv()
	env["FILVAULT_JWT_SECRET"] = "too-short"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for short JWT in production")
	}
	if !strings.Contains(err.Error(), "FILVAULT_JWT_SECRET") {
		t.Fatalf("error = %v, want JWT length guidance", err)
	}
}

func TestLoad_ProductionRejectsDatabaseSSLDisable(t *testing.T) {
	env := baseProdEnv()
	env["FILVAULT_DATABASE_URL"] = "postgres://app:strong-db-pass@db.example.com:5432/filvault?sslmode=disable"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for sslmode=disable in production")
	}
	if !strings.Contains(err.Error(), "FILVAULT_DATABASE_URL") {
		t.Fatalf("error = %v, want database SSL guidance", err)
	}
}

func TestLoad_ProductionRejectsDatabaseWithoutSSLMode(t *testing.T) {
	env := baseProdEnv()
	env["FILVAULT_DATABASE_URL"] = "postgres://app:strong-db-pass@db.example.com:5432/filvault"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when sslmode is omitted in production")
	}
	if !strings.Contains(err.Error(), "sslmode") {
		t.Fatalf("error = %v, want sslmode guidance", err)
	}
}

func TestLoad_ProductionRejectsPlaceholderS3Secret(t *testing.T) {
	env := baseProdEnv()
	env["FILVAULT_S3_SECRET_KEY"] = "replace-with-local-s3-secret-key"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for placeholder S3 secret in production")
	}
	if !strings.Contains(err.Error(), "FILVAULT_S3_SECRET_KEY") {
		t.Fatalf("error = %v, want S3 secret guidance", err)
	}
}

func TestLoad_ProductionRejectsHistoricalDefaultJWT(t *testing.T) {
	env := baseProdEnv()
	env["FILVAULT_JWT_SECRET"] = "dev-jwt-secret"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for historical dev JWT default in production")
	}
}

func TestLoad_ProductionRejectsDevSeedUser(t *testing.T) {
	env := baseProdEnv()
	env["FILVAULT_SEED_DEV_USER"] = "true"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when dev seed enabled in production")
	}
	if !strings.Contains(err.Error(), "FILVAULT_SEED_DEV_USER") {
		t.Fatalf("error = %v, want seed dev user guidance", err)
	}
}

func TestLoad_ProductionRejectsLocalhostCORS(t *testing.T) {
	env := baseProdEnv()
	env["FILVAULT_CORS_ALLOWED_ORIGINS"] = "http://localhost:5173"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for localhost CORS in production")
	}
	if !strings.Contains(err.Error(), "FILVAULT_CORS_ALLOWED_ORIGINS") {
		t.Fatalf("error = %v, want CORS guidance", err)
	}
}

func TestIsUnsafeSecret(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"replace-with-local-jwt-secret-min-32-chars", true},
		{"dev-jwt-secret", true},
		{"dev-invite", true},
		{"filvaultsecret", true},
		{"devkey", true},
		{"secret", true},
		{"unique-local-only-value-chosen-by-developer", false},
	}
	for _, tc := range cases {
		if got := isUnsafeSecret(tc.value); got != tc.want {
			t.Errorf("isUnsafeSecret(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
