package sharelink

import (
	"context"
	"time"
)

// Repository persists share links. TokenHash is the only stored form of the
// token; plaintext exists only in the create response.
type Repository interface {
	// Create inserts a new link row (already revoked or active).
	Create(ctx context.Context, link ShareLink) error
	// CountActiveByOwner returns how many active links the user holds.
	CountActiveByOwner(ctx context.Context, ownerID string) (int, error)
	// FileIsShareable reports whether the file is owned, READY and not trashed.
	FileIsShareable(ctx context.Context, ownerID, fileID string) (bool, error)
	// RevokeActiveForFile revokes every active link of the file.
	RevokeActiveForFile(ctx context.Context, fileID string, at time.Time) error
	// ListActiveByOwner returns active links with joined file info, newest first.
	ListActiveByOwner(ctx context.Context, ownerID string) ([]ShareLink, error)
	// FindActiveByTokenHash resolves a token hash to an active link whose
	// file is still READY and not trashed. No match → (nil, nil).
	FindActiveByTokenHash(ctx context.Context, tokenHash string) (*ShareLink, error)
	// FileName returns the display name of an owned file ("" when missing).
	FileName(ctx context.Context, ownerID, fileID string) (string, error)
}
