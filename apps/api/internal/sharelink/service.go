package sharelink

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"filvault/internal/activity"
	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/platform/config"
	"filvault/internal/platform/objectstore"
)

// Objects presigns downloads for shared files.
type Objects interface {
	CreateDownloadURL(ctx context.Context, key string, opts objectstore.DownloadOptions) (objectstore.PresignedURL, error)
}

type Service struct {
	repo     Repository
	objects  Objects
	activity ActivityRecorder
}

// ActivityRecorder records share lifecycle events; failures never break shares.
type ActivityRecorder interface {
	Record(ctx context.Context, ownerID, eventType, targetName string)
}

func NewService(repo Repository, objects Objects, activity ActivityRecorder) *Service {
	return &Service{repo: repo, objects: objects, activity: activity}
}

type Created struct {
	ID        string     `json:"id"`
	FileID    string     `json:"fileId"`
	URL       string     `json:"url"`
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

// Create mints a fresh public link for an owned, READY file. Any existing
// active link of the same file is revoked first (one link per file).
func (s *Service) Create(ctx context.Context, ownerID, fileID, expiresIn string) (*Created, error) {
	ttl, ok := allowedExpiresIn[expiresIn]
	if expiresIn != "" && !ok {
		return nil, apperr.Validation
	}

	shareable, err := s.repo.FileIsShareable(ctx, ownerID, fileID)
	if err != nil {
		return nil, err
	}
	if !shareable {
		return nil, apperr.NotFound
	}

	active, err := s.repo.CountActiveByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if active >= DefaultActiveLinksPerUser {
		return nil, apperr.Conflict
	}

	token, tokenHash, err := newToken()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var expiresAt *time.Time
	if ttl > 0 {
		at := now.Add(ttl)
		expiresAt = &at
	}

	link := ShareLink{
		ID:        auth.NewID(),
		OwnerID:   ownerID,
		FileID:    fileID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
	if err := s.repo.RevokeActiveForFile(ctx, fileID, now); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, link); err != nil {
		return nil, err
	}
	fileName, err := s.fileName(ctx, ownerID, fileID)
	if err != nil {
		return nil, err
	}
	s.activity.Record(ctx, ownerID, activity.TypeShareCreated, fileName)

	return &Created{
		ID:        link.ID,
		FileID:    fileID,
		URL:       publicPath(token),
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

// Revoke disables every active link of the file. Unknown/foreign → NOT_FOUND.
func (s *Service) Revoke(ctx context.Context, ownerID, fileID string) error {
	shareable, err := s.repo.FileIsShareable(ctx, ownerID, fileID)
	if err != nil {
		return err
	}
	if !shareable {
		return apperr.NotFound
	}
	if err := s.repo.RevokeActiveForFile(ctx, fileID, time.Now().UTC()); err != nil {
		return err
	}
	fileName, err := s.fileName(ctx, ownerID, fileID)
	if err != nil {
		return err
	}
	s.activity.Record(ctx, ownerID, activity.TypeShareRevoked, fileName)
	return nil
}

// List returns the user's active links, newest first.
func (s *Service) List(ctx context.Context, ownerID string) ([]ShareLink, error) {
	return s.repo.ListActiveByOwner(ctx, ownerID)
}

// PublicMeta serves anonymous metadata; any failure collapses to NOT_FOUND.
func (s *Service) PublicMeta(ctx context.Context, token string) (*PublicMeta, error) {
	link, err := s.resolve(ctx, token)
	if err != nil {
		return nil, err
	}
	return &PublicMeta{
		Name:      link.FileInfo.Name,
		MimeType:  link.FileInfo.MimeType,
		SizeBytes: link.FileInfo.SizeBytes,
		ExpiresAt: link.ExpiresAt,
	}, nil
}

type PublicDownload struct {
	URL       string    `json:"downloadUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// PublicDownload presigns a short-lived download URL; failure → NOT_FOUND.
func (s *Service) PublicDownload(ctx context.Context, token string) (*PublicDownload, error) {
	link, err := s.resolve(ctx, token)
	if err != nil {
		return nil, err
	}
	presigned, err := s.objects.CreateDownloadURL(ctx, link.FileInfo.ObjectKey, objectstore.DownloadOptions{
		Expires: config.DownloadPresignTTL,
	})
	if err != nil {
		return nil, err
	}
	return &PublicDownload{URL: presigned.URL, ExpiresAt: presigned.ExpiresAt}, nil
}

func (s *Service) resolve(ctx context.Context, token string) (*ShareLink, error) {
	if len(token) != tokenHexLen {
		return nil, apperr.NotFound
	}
	link, err := s.repo.FindActiveByTokenHash(ctx, hashToken(token))
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, apperr.NotFound
	}
	return link, nil
}

// publicPath is the share page route served by the web app; the API returns
// it as a relative URL so the caller joins its own origin.
func publicPath(token string) string {
	return "/s/" + token
}

// fileName resolves the display name of an owned file for activity logging.
func (s *Service) fileName(ctx context.Context, ownerID, fileID string) (string, error) {
	return s.repo.FileName(ctx, ownerID, fileID)
}

func newToken() (plaintext, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	plaintext = hex.EncodeToString(buf)
	return plaintext, hashToken(plaintext), nil
}

func hashToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}
