package config

import "time"

const (
	DefaultStorageQuotaBytes  int64 = 10 * 1024 * 1024 * 1024
	DefaultTrashRetentionDays       = 30
	MinPasswordLength               = 8
	MaxFileSizeBytes          int64 = 104857600 // 100 MiB
	UploadPresignTTL                = 15 * time.Minute
	DownloadPresignTTL              = 5 * time.Minute

	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

type Config struct {
	DatabaseURL               string
	InviteCode                string
	JWTSecret                 string
	HTTPAddr                  string
	MetadataStore             string
	Mailer                    string
	SMTPHost                  string
	SMTPPort                  string
	SMTPUsername              string
	SMTPPassword              string
	SMTPFrom                  string
	ObjectStore               string
	S3Endpoint                string
	S3Bucket                  string
	S3Region                  string
	S3AccessKey               string
	S3SecretKey               string
	S3PublicEndpoint          string
	DefaultTrashAutoDelete    bool
	DefaultTrashRetentionDays int
	CORSAllowedOrigins        []string
}
