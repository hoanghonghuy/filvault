package config

import (
	"fmt"
	"net/url"
)

type Env string

const (
	EnvDevelopment Env = "development"
	EnvProduction  Env = "production"
)

const minProductionJWTLength = 32

// unsafeSecretValues lists historically committed or documented placeholder values
// that must never be used outside local development.
var unsafeSecretValues = []string{
	"replace-with-local-postgres-password",
	"replace-with-local-s3-secret-key",
	"replace-with-local-jwt-secret-min-32-chars",
	"replace-with-local-invite-code",
	"replace-with-local-seed-password",
	"replace-with-local-livekit-api-key",
	"replace-with-local-livekit-api-secret",
	"dev-jwt-secret",
	"dev-invite",
	"filvaultsecret",
	"devkey",
	"secret",
	"filvault",
}

func parseEnv(raw string) (Env, error) {
	trimmed := stringsTrim(raw)
	if trimmed == "" {
		return EnvDevelopment, nil
	}
	switch stringsToLower(trimmed) {
	case "production", "prod":
		return EnvProduction, nil
	case "development", "dev":
		return EnvDevelopment, nil
	default:
		return "", fmt.Errorf("FILVAULT_ENV must be development or production (got %q)", trimmed)
	}
}

func isUnsafeSecret(value string) bool {
	normalized := stringsToLower(stringsTrim(value))
	if normalized == "" {
		return false
	}
	for _, blocked := range unsafeSecretValues {
		if normalized == blocked {
			return true
		}
	}
	if stringsHasPrefix(normalized, "replace-with-") {
		return true
	}
	return false
}

func validateProduction(cfg Config) error {
	if err := requireNonUnsafe("FILVAULT_JWT_SECRET", cfg.JWTSecret); err != nil {
		return err
	}
	if len(cfg.JWTSecret) < minProductionJWTLength {
		return fmt.Errorf("FILVAULT_JWT_SECRET must be at least %d characters in production", minProductionJWTLength)
	}

	if err := validateProductionDatabaseURL(cfg.DatabaseURL); err != nil {
		return err
	}

	if cfg.InviteCode == "" {
		return fmt.Errorf("FILVAULT_INVITE_CODE is required in production")
	}
	if err := requireNonUnsafe("FILVAULT_INVITE_CODE", cfg.InviteCode); err != nil {
		return err
	}
	if len(cfg.InviteCode) < 8 {
		return fmt.Errorf("FILVAULT_INVITE_CODE must be at least 8 characters in production")
	}

	if cfg.ObjectStore == "s3" {
		if err := requireNonUnsafe("FILVAULT_S3_ACCESS_KEY", cfg.S3AccessKey); err != nil {
			return err
		}
		if err := requireNonUnsafe("FILVAULT_S3_SECRET_KEY", cfg.S3SecretKey); err != nil {
			return err
		}
		if stringsTrim(cfg.S3PublicEndpoint) == "" {
			return fmt.Errorf("FILVAULT_S3_PUBLIC_ENDPOINT is required in production when using S3 object storage")
		}
	}

	for _, origin := range cfg.CORSAllowedOrigins {
		if isLocalHostOrigin(origin) {
			return fmt.Errorf("FILVAULT_CORS_ALLOWED_ORIGINS must not include localhost or 127.0.0.1 in production (got %q)", origin)
		}
	}

	if err := validateProductionMailer(cfg); err != nil {
		return err
	}

	if err := validateProductionLiveKit(cfg); err != nil {
		return err
	}

	if stringsEqualFold(osGetenv("FILVAULT_SEED_DEV_USER"), "true") {
		return fmt.Errorf("FILVAULT_SEED_DEV_USER must not be true in production")
	}

	return nil
}

func validateProductionDatabaseURL(raw string) error {
	mode, err := databaseSSLMode(raw)
	if err != nil {
		return fmt.Errorf("FILVAULT_DATABASE_URL is invalid: %w", err)
	}
	switch mode {
	case "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("FILVAULT_DATABASE_URL sslmode must be require, verify-ca, or verify-full in production (got %q)", mode)
	}
	host, err := databaseHost(raw)
	if err != nil {
		return fmt.Errorf("FILVAULT_DATABASE_URL is invalid: %w", err)
	}
	if isLocalDatabaseHost(host) {
		return fmt.Errorf("FILVAULT_DATABASE_URL must not point to localhost or 127.0.0.1 in production")
	}
	return nil
}

func databaseSSLMode(raw string) (string, error) {
	u, err := urlParse(raw)
	if err != nil {
		return "", err
	}
	value, err := queryParamSingle(u, "sslmode")
	if err != nil {
		return "", err
	}
	return stringsToLower(value), nil
}

func queryParamSingle(u *url.URL, key string) (string, error) {
	vals, ok := u.Query()[key]
	if !ok || len(vals) == 0 {
		return "", fmt.Errorf("%s query parameter is required", key)
	}
	if len(vals) > 1 {
		return "", fmt.Errorf("%s query parameter must appear once", key)
	}
	value := stringsTrim(vals[0])
	if value == "" {
		return "", fmt.Errorf("%s query parameter must not be empty", key)
	}
	return value, nil
}

func validateProductionMailer(cfg Config) error {
	mode := stringsToLower(stringsTrim(cfg.Mailer))
	if mode == "" || mode == "auto" || mode == "console" {
		return fmt.Errorf("FILVAULT_MAILER must be smtp in production; console logging is not allowed")
	}
	if stringsTrim(cfg.SMTPHost) == "" {
		return fmt.Errorf("FILVAULT_SMTP_HOST is required in production")
	}
	if stringsTrim(cfg.SMTPFrom) == "" {
		return fmt.Errorf("FILVAULT_SMTP_FROM is required in production")
	}
	return nil
}

func validateProductionLiveKit(cfg Config) error {
	key := stringsTrim(cfg.LiveKitAPIKey)
	secret := stringsTrim(cfg.LiveKitAPISecret)
	if key == "" && secret == "" {
		return nil
	}
	if key == "" || secret == "" {
		return fmt.Errorf("FILVAULT_LIVEKIT_API_KEY and FILVAULT_LIVEKIT_API_SECRET must both be set in production when LiveKit is enabled")
	}
	if err := requireNonUnsafe("FILVAULT_LIVEKIT_API_KEY", key); err != nil {
		return err
	}
	if err := requireNonUnsafe("FILVAULT_LIVEKIT_API_SECRET", secret); err != nil {
		return err
	}
	if isLocalHostOrigin(cfg.LiveKitPublicURL) || stringsContains(stringsToLower(cfg.LiveKitPublicURL), "localhost") {
		return fmt.Errorf("FILVAULT_LIVEKIT_PUBLIC_URL must not use localhost in production")
	}
	return nil
}

func requireNonUnsafe(name, value string) error {
	if isUnsafeSecret(value) {
		return fmt.Errorf("%s uses a documented placeholder or historical default; set a unique production value via your secret manager", name)
	}
	return nil
}

func isLocalHostOrigin(raw string) bool {
	lower := stringsToLower(stringsTrim(raw))
	return stringsContains(lower, "localhost") || stringsContains(lower, "127.0.0.1")
}

func isLocalDatabaseHost(host string) bool {
	switch stringsToLower(stringsTrim(host)) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

func databaseHost(raw string) (string, error) {
	u, err := urlParse(raw)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}
