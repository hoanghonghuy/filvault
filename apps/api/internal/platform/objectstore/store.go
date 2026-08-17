package objectstore

import (
	"context"
	"time"
)

type PresignedURL struct {
	URL       string
	ExpiresAt time.Time
}

type ObjectStat struct {
	Size        int64
	ContentType string
}

type UploadOptions struct {
	ContentType string
	Expires     time.Duration
}

type DownloadOptions struct {
	Expires time.Duration
}

type ObjectStore interface {
	CreateUploadURL(ctx context.Context, key string, opts UploadOptions) (PresignedURL, error)
	CreateDownloadURL(ctx context.Context, key string, opts DownloadOptions) (PresignedURL, error)
	Head(ctx context.Context, key string) (ObjectStat, error)
	Delete(ctx context.Context, key string) error
}
