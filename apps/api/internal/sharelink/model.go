package sharelink

import "time"

// FileInfo carries the joined file columns the feature needs.
type FileInfo struct {
	ID        string
	Name      string
	MimeType  string
	SizeBytes int64
	ObjectKey string
	Status    string
}

// ShareLink is a public link row. Token plaintext never reaches the store:
// only SHA-256 hashes are persisted (ADR 0004).
type ShareLink struct {
	ID           string
	OwnerID      string
	FileID       string
	TokenHash    string
	ExpiresAt    *time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
	FileInfo     FileInfo
	RecipientUID *string
}

// PublicMeta is the minimal payload served to anonymous visitors.
type PublicMeta struct {
	Name      string     `json:"name"`
	MimeType  string     `json:"mimeType"`
	SizeBytes int64      `json:"sizeBytes"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

const (
	// DefaultActiveLinksPerUser caps how many live links one user may hold.
	DefaultActiveLinksPerUser = 20
	// tokenHexLen is the hex encoding of the 32-byte random token.
	tokenHexLen = 64
)

// allowedExpiresIn accepts only the spec'd TTL choices; empty means never expires.
var allowedExpiresIn = map[string]time.Duration{
	"1h":  1 * time.Hour,
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
}
