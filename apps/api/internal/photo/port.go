package photo

import (
	"context"
	"time"
)

type Repository interface {
	ListTimelineFiles(ctx context.Context, ownerID string, before *time.Time, limit int) ([]TimelineItem, error)
	GetPhotoFile(ctx context.Context, ownerID, fileID string) (*TimelineItem, error)

	CreateAlbum(ctx context.Context, a Album) error
	GetAlbum(ctx context.Context, ownerID, id string) (*Album, error)
	ListAlbums(ctx context.Context, ownerID string) ([]Album, error)
	UpdateAlbum(ctx context.Context, a Album) error
	DeleteAlbum(ctx context.Context, ownerID, id string) error

	AddAlbumItem(ctx context.Context, albumID, fileID string, position int, at time.Time) error
	RemoveAlbumItem(ctx context.Context, albumID, fileID string) error
	ListAlbumItems(ctx context.Context, ownerID, albumID string) ([]TimelineItem, error)
}
