package folder

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, f Folder) error
	GetAliveByID(ctx context.Context, ownerID, id string) (*Folder, error)
	Update(ctx context.Context, f Folder) error
	SoftDelete(ctx context.Context, ownerID, id string, at time.Time) error
	ListAliveChildren(ctx context.Context, ownerID string, parentID *string) ([]Folder, error)
	CountAliveChildren(ctx context.Context, ownerID, folderID string) (subfolders int, files int, err error)
	ListAliveFilesInFolder(ctx context.Context, ownerID string, folderID *string) ([]BrowserFile, error)
	IsAncestor(ctx context.Context, ownerID, ancestorID, nodeID string) (bool, error)
}
