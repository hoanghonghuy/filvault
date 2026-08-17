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
	ExistsAliveByName(ctx context.Context, ownerID string, folderID *string, name string) (bool, error)
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
