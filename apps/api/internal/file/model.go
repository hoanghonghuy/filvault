package file

import "time"

const (
	StatusPending = "PENDING"
	StatusReady   = "READY"
	StatusFailed  = "FAILED"
)

type File struct {
	ID              string
	OwnerID         string
	FolderID        *string
	Name            string
	OriginalName    string
	ObjectKey       string
	MimeType        string
	SizeBytes       int64
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	UploadExpiresAt *time.Time
	ReplacesFileID  *string
}

// FileVersion is an archived previous version of a file.
type FileVersion struct {
	ID        string
	FileID    string
	ObjectKey string
	SizeBytes int64
	MimeType  string
	CreatedAt time.Time
}

// FavoriteFile is a favorited file rendered in the favorites list.
type FavoriteFile struct {
	ID          string
	Name        string
	MimeType    string
	SizeBytes   int64
	UpdatedAt   time.Time
	FavoritedAt time.Time
}

const (
	DefaultFavoritesLimit = 50
	MaxFavoritesLimit     = 100
)

func ObjectKey(ownerID, fileID string) string {
	return "users/" + ownerID + "/files/" + fileID
}
