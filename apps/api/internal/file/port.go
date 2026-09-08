package file

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, f File) error
	GetByID(ctx context.Context, ownerID, id string) (*File, error)
	Update(ctx context.Context, f File) error
	DeleteRow(ctx context.Context, ownerID, id string) error
	SoftDelete(ctx context.Context, ownerID, id string, at time.Time) error
	ExistsAliveByName(ctx context.Context, ownerID string, folderID *string, name, excludeFileID string) (bool, error)
	GetAliveByName(ctx context.Context, ownerID string, folderID *string, name string) (*File, error)

	CreateVersion(ctx context.Context, v FileVersion) error
	ListVersions(ctx context.Context, fileID string) ([]FileVersion, error)
	GetVersion(ctx context.Context, fileID, versionID string) (*FileVersion, error)

	AddFavorite(ctx context.Context, ownerID, fileID string, at time.Time) error
	RemoveFavorite(ctx context.Context, ownerID, fileID string) error
	ListFavorites(ctx context.Context, ownerID string, limit int) ([]FavoriteFile, error)

	CompleteUpload(ctx context.Context, ownerID, fileID string, size int64, contentType string, now time.Time) (*File, error)
	CompleteReplaceUpload(ctx context.Context, ownerID, pendingFileID, targetFileID string, size int64, contentType string, now time.Time) (*File, error)
}

type QuotaStore interface {
	GetStorage(ctx context.Context, userID string) (used, quota int64, err error)
	AddStorageUsed(ctx context.Context, userID string, delta int64) error
}

type UploadSession struct {
	FileID    string
	UploadURL string
	ExpiresAt time.Time
}

type DownloadURL struct {
	URL       string
	ExpiresAt time.Time
}
