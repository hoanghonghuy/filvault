package share

import (
	"context"
	"time"

	"filvault/internal/file"
	"filvault/internal/folder"
	"filvault/internal/platform/objectstore"
	"filvault/internal/user"
)

type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	GetUserByID(ctx context.Context, id string) (*user.User, error)
	GetAliveFileByID(ctx context.Context, ownerID, id string) (*file.File, error)
	GetAliveFolderByID(ctx context.Context, ownerID, id string) (*folder.Folder, error)
	GetFileByIDAny(ctx context.Context, id string) (*file.File, error)

	CreateShare(ctx context.Context, s Share) error
	ListSharesByOwner(ctx context.Context, ownerID string) ([]Share, error)
	ListSharesByRecipient(ctx context.Context, recipientID string) ([]Share, error)
	DeleteShare(ctx context.Context, ownerID, id string) error
	GetShareByRecipientResource(ctx context.Context, recipientID, resourceType, resourceID string) (*Share, error)

	ListAliveFolderChildren(ctx context.Context, ownerID string, parentID *string) ([]folder.Folder, error)
	ListAliveFilesInFolder(ctx context.Context, ownerID string, folderID *string) ([]folder.BrowserFile, error)
}

type ObjectStore interface {
	CreateDownloadURL(ctx context.Context, key string, opts objectstore.DownloadOptions) (objectstore.PresignedURL, error)
}

// DownloadURL is the presigned download result for a shared file.
type DownloadURL struct {
	URL       string
	ExpiresAt time.Time
}
