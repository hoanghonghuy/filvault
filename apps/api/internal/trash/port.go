package trash

import (
	"context"
	"time"

	"filvault/internal/file"
	"filvault/internal/folder"
	"filvault/internal/user"
)

type Repository interface {
	GetAnyFile(ctx context.Context, ownerID, id string) (*file.File, error)
	SoftDeleteFile(ctx context.Context, ownerID, id string, at time.Time) error
	RestoreFile(ctx context.Context, ownerID, id string) error
	PurgeFile(ctx context.Context, ownerID, id string) (objectKey string, err error)
	CompleteFilePurge(ctx context.Context, ownerID, fileID string) error
	ClaimPurgeJob(ctx context.Context, workerID string, lease time.Duration) (*PurgeJob, error)
	CompletePurgeJob(ctx context.Context, jobID, workerID string) error
	FailPurgeJob(ctx context.Context, jobID, workerID, lastError string, retryAt time.Time, dead bool) error
	ClaimUploadCleanupJob(ctx context.Context, workerID string, lease time.Duration) (*UploadCleanupJob, error)
	CompleteUploadCleanupJob(ctx context.Context, jobID, workerID, ownerID, fileID string) error
	FailUploadCleanupJob(ctx context.Context, jobID, workerID, lastError string, retryAt time.Time, dead bool) error
	QueueExpiredUploadCleanupJobs(ctx context.Context, now time.Time) error
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

type PurgeJob struct {
	ID        string
	OwnerID   string
	FileID    string
	ObjectKey string
	Attempts  int
}

type UploadCleanupJob struct {
	ID        string
	OwnerID   string
	FileID    string
	ObjectKey string
	Attempts  int
}

type QuotaStore interface {
	AddStorageUsed(ctx context.Context, userID string, delta int64) error
}

type ObjectStore interface {
	Delete(ctx context.Context, key string) error
}
