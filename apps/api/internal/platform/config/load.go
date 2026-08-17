package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:               os.Getenv("FILNEST_DATABASE_URL"),
		InviteCode:                os.Getenv("FILNEST_INVITE_CODE"),
		JWTSecret:                 os.Getenv("FILNEST_JWT_SECRET"),
		HTTPAddr:                  getenv("FILNEST_HTTP_ADDR", ":8080"),
		MetadataStore:             getenv("FILNEST_METADATA_STORE", "postgres"),
		Mailer:                    getenv("FILNEST_MAILER", "auto"),
		SMTPHost:                  os.Getenv("FILNEST_SMTP_HOST"),
		SMTPPort:                  getenv("FILNEST_SMTP_PORT", "587"),
		SMTPUsername:              os.Getenv("FILNEST_SMTP_USERNAME"),
		SMTPPassword:              os.Getenv("FILNEST_SMTP_PASSWORD"),
		SMTPFrom:                  os.Getenv("FILNEST_SMTP_FROM"),
		DefaultTrashAutoDelete:    getenv("FILNEST_DEFAULT_TRASH_AUTO_DELETE", "false") == "true",
		DefaultTrashRetentionDays: DefaultTrashRetentionDays,
	}
	if days := os.Getenv("FILNEST_DEFAULT_TRASH_RETENTION_DAYS"); days != "" {
		n, err := strconv.Atoi(days)
		if err != nil || n < 1 {
			return Config{}, fmt.Errorf("FILNEST_DEFAULT_TRASH_RETENTION_DAYS must be >= 1")
		}
		cfg.DefaultTrashRetentionDays = n
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("FILNEST_DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("FILNEST_JWT_SECRET is required")
	}
	if strings.EqualFold(cfg.MetadataStore, "dynamodb") {
		return Config{}, fmt.Errorf("FILNEST_METADATA_STORE=dynamodb is not implemented")
	}
	if cfg.MetadataStore != "postgres" {
		return Config{}, fmt.Errorf("unsupported FILNEST_METADATA_STORE %q", cfg.MetadataStore)
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
