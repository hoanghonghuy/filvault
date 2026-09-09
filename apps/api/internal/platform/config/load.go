package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:                os.Getenv("FILVAULT_DATABASE_URL"),
		InviteCode:                 os.Getenv("FILVAULT_INVITE_CODE"),
		JWTSecret:                  os.Getenv("FILVAULT_JWT_SECRET"),
		HTTPAddr:                   getenv("FILVAULT_HTTP_ADDR", ":8080"),
		MetadataStore:              getenv("FILVAULT_METADATA_STORE", "postgres"),
		Mailer:                     getenv("FILVAULT_MAILER", "auto"),
		SMTPHost:                   os.Getenv("FILVAULT_SMTP_HOST"),
		SMTPPort:                   getenv("FILVAULT_SMTP_PORT", "587"),
		SMTPUsername:               os.Getenv("FILVAULT_SMTP_USERNAME"),
		SMTPPassword:               os.Getenv("FILVAULT_SMTP_PASSWORD"),
		SMTPFrom:                   os.Getenv("FILVAULT_SMTP_FROM"),
		ObjectStore:                getenv("FILVAULT_OBJECT_STORE", "s3"),
		S3Endpoint:                 os.Getenv("FILVAULT_S3_ENDPOINT"),
		S3Bucket:                   getenv("FILVAULT_S3_BUCKET", "filvault"),
		S3Region:                   getenv("FILVAULT_S3_REGION", "us-east-1"),
		S3AccessKey:                os.Getenv("FILVAULT_S3_ACCESS_KEY"),
		S3SecretKey:                os.Getenv("FILVAULT_S3_SECRET_KEY"),
		S3PublicEndpoint:           os.Getenv("FILVAULT_S3_PUBLIC_ENDPOINT"),
		DefaultImageThumbnails:     getenv("FILVAULT_DEFAULT_IMAGE_THUMBNAILS", "true") == "true",
		DefaultVideoThumbnails:     getenv("FILVAULT_DEFAULT_VIDEO_THUMBNAILS", "true") == "true",
		DefaultTrashAutoDelete:     getenv("FILVAULT_DEFAULT_TRASH_AUTO_DELETE", "false") == "true",
		DefaultTrashRetentionDays:  DefaultTrashRetentionDays,
		PublicShareRateLimitPerMin: 30,
		CORSAllowedOrigins:         parseCSV(getenv("FILVAULT_CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
		LiveKitURL:                 getenv("FILVAULT_LIVEKIT_URL", "http://localhost:7880"),
		LiveKitPublicURL:           getenv("FILVAULT_LIVEKIT_PUBLIC_URL", "ws://localhost:7880"),
		LiveKitAPIKey:              os.Getenv("FILVAULT_LIVEKIT_API_KEY"),
		LiveKitAPISecret:           os.Getenv("FILVAULT_LIVEKIT_API_SECRET"),
	}
	if days := os.Getenv("FILVAULT_DEFAULT_TRASH_RETENTION_DAYS"); days != "" {
		n, err := strconv.Atoi(days)
		if err != nil || n < 1 {
			return Config{}, fmt.Errorf("FILVAULT_DEFAULT_TRASH_RETENTION_DAYS must be >= 1")
		}
		cfg.DefaultTrashRetentionDays = n
	}
	if raw := os.Getenv("FILVAULT_PUBLIC_SHARE_RATE_LIMIT_PER_MIN"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return Config{}, fmt.Errorf("FILVAULT_PUBLIC_SHARE_RATE_LIMIT_PER_MIN must be >= 1")
		}
		cfg.PublicShareRateLimitPerMin = n
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("FILVAULT_DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("FILVAULT_JWT_SECRET is required")
	}
	if cfg.ObjectStore == "s3" && (cfg.S3Endpoint == "" || cfg.S3AccessKey == "" || cfg.S3SecretKey == "") {
		return Config{}, fmt.Errorf("FILVAULT_S3_ENDPOINT, FILVAULT_S3_ACCESS_KEY and FILVAULT_S3_SECRET_KEY are required for S3 object storage")
	}
	if strings.EqualFold(cfg.MetadataStore, "dynamodb") {
		return Config{}, fmt.Errorf("FILVAULT_METADATA_STORE=dynamodb is not implemented")
	}
	if cfg.MetadataStore != "postgres" {
		return Config{}, fmt.Errorf("unsupported FILVAULT_METADATA_STORE %q", cfg.MetadataStore)
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
