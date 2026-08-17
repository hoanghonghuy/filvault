package trash

import (
	"context"
	"time"

	"filnest/internal/file"
	"filnest/internal/folder"
	"filnest/internal/user"
)

type Repository interface {
	GetAnyFile(ctx context.Context, ownerID, id string) (*file.File, error)
	SoftDeleteFile(ctx context.Context, ownerID, id string, at time.Time) error
	RestoreFile(ctx context.Context, ownerID, id string) error
	DeleteFileRow(ctx context.Context, ownerID, id string) error
	ListTrashedFiles(ctx context.Context, ownerID string) ([]file.File, error)
	ListExpiredTrashedFiles(ctx context.Context, ownerID string, deletedBefore time.Time) ([]file.File, error)
	ExistsAliveFileByName(ctx context.Context, ownerID string, folderID *string, name, excludeFileID string) (bool, error)

	GetAnyFolder(ctx context.Context, ownerID, id string) (*folder.Folder, error)
	GetAliveFolderByID(ctx context.Context, ownerID, id string) (*folder.Folder, error)
	RestoreFolder(ctx context.Context, ownerID, id string) error
	DeleteFolderRow(ctx context.Context, ownerID, id string) error
	ListTrashedFolders(ctx context.Context, ownerID string) ([]folder.Folder, error)
	ListExpiredTrashedFolders(ctx context.Context, ownerID string, deletedBefore time.Time) ([]folder.Folder, error)
	CountAllFolderChildren(ctx context.Context, ownerID, folderID string) (subfolders int, files int, err error)
	ExistsAliveFolderByName(ctx context.Context, ownerID string, parentID *string, name, excludeFolderID string) (bool, error)

	GetUserByID(ctx context.Context, id string) (*user.User, error)
	ListUsersWithAutoTrash(ctx context.Context) ([]user.User, error)
	UpdateTrashSettings(ctx context.Context, userID string, enabled bool, retentionDays int) error
}

type QuotaStore interface {
	AddStorageUsed(ctx context.Context, userID string, delta int64) error
}

type ObjectStore interface {
	Delete(ctx context.Context, key string) error
}
