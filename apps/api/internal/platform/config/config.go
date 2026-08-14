package config

const (
	DefaultStorageQuotaBytes  int64 = 10 * 1024 * 1024 * 1024
	DefaultTrashRetentionDays       = 30
	MinPasswordLength               = 8
)

type Config struct {
	DatabaseURL               string
	InviteCode                string
	JWTSecret                 string
	HTTPAddr                  string
	MetadataStore             string
	DefaultTrashAutoDelete    bool
	DefaultTrashRetentionDays int
}
