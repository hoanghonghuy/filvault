package file

import (
	"context"
	"errors"
	"strings"
	"time"

	"filvault/internal/activity"
	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/folder"
	"filvault/internal/platform/config"
	"filvault/internal/platform/objectstore"
)

type Service struct {
	repo     Repository
	folders  folder.Repository
	quota    QuotaStore
	objects  objectstore.ObjectStore
	activity ActivityRecorder
	now      func() time.Time
}

// ActivityRecorder records lifecycle events; failures never break uploads.
type ActivityRecorder interface {
	Record(ctx context.Context, ownerID, eventType, targetName string)
}

func NewService(repo Repository, folders folder.Repository, quota QuotaStore, objects objectstore.ObjectStore, activity ActivityRecorder) *Service {
	return &Service{
		repo:     repo,
		folders:  folders,
		quota:    quota,
		objects:  objects,
		activity: activity,
		now:      time.Now,
	}
}

func (s *Service) CreateUploadSession(ctx context.Context, ownerID, name, contentType string, size int64, folderID *string, replaceFileID *string) (UploadSession, error) {
	if err := ValidateUpload(name, contentType, size); err != nil {
		return UploadSession{}, err
	}
	if folderID != nil && *folderID == "" {
		folderID = nil
	}
	if replaceFileID != nil && *replaceFileID == "" {
		replaceFileID = nil
	}
	if folderID != nil {
		f, err := s.folders.GetAliveByID(ctx, ownerID, *folderID)
		if err != nil {
			return UploadSession{}, err
		}
		if f == nil {
			return UploadSession{}, apperr.NotFound
		}
	}
	used, quota, err := s.quota.GetStorage(ctx, ownerID)
	if err != nil {
		return UploadSession{}, err
	}
	if used+size > quota {
		return UploadSession{}, apperr.QuotaExceeded
	}
	displayName := strings.TrimSpace(name)

	if replaceFileID != nil {
		target, err := s.repo.GetByID(ctx, ownerID, *replaceFileID)
		if err != nil {
			return UploadSession{}, err
		}
		if target == nil || target.DeletedAt != nil || target.Status != StatusReady {
			return UploadSession{}, apperr.NotFound
		}
		if target.Name != displayName {
			return UploadSession{}, apperr.Validation
		}
	} else {
		existing, err := s.repo.GetAliveByName(ctx, ownerID, folderID, displayName)
		if err != nil {
			return UploadSession{}, err
		}
		if existing != nil {
			if existing.Status == StatusReady {
				return UploadSession{}, apperr.Conflict
			}
			// Clean up previous incomplete/abandoned/failed upload session for this name
			_ = s.objects.Delete(ctx, existing.ObjectKey)
			_ = s.repo.DeleteRow(ctx, ownerID, existing.ID)
		}
	}

	now := s.now().UTC()
	fileID := auth.NewID()
	key := ObjectKey(ownerID, fileID)
	expires := now.Add(config.UploadPresignTTL)

	f := File{
		ID:              fileID,
		OwnerID:         ownerID,
		FolderID:        folderID,
		Name:            displayName,
		OriginalName:    displayName,
		ObjectKey:       key,
		MimeType:        contentType,
		SizeBytes:       size,
		Status:          StatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
		UploadExpiresAt: &expires,
		ReplacesFileID:  replaceFileID,
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return UploadSession{}, err
	}

	presigned, err := s.objects.CreateUploadURL(ctx, key, objectstore.UploadOptions{
		ContentType: contentType,
		Expires:     config.UploadPresignTTL,
	})
	if err != nil {
		_ = s.repo.DeleteRow(ctx, ownerID, fileID)
		return UploadSession{}, err
	}
	return UploadSession{
		FileID:    fileID,
		UploadURL: presigned.URL,
		ExpiresAt: presigned.ExpiresAt,
	}, nil
}

func (s *Service) Complete(ctx context.Context, ownerID, fileID string) (File, error) {
	f, err := s.getPending(ctx, ownerID, fileID)
	if err != nil {
		return File{}, err
	}
	now := s.now().UTC()
	if f.UploadExpiresAt != nil && !now.Before(*f.UploadExpiresAt) {
		_ = s.markFailed(ctx, *f)
		return File{}, apperr.UploadExpired
	}

	stat, err := s.objects.Head(ctx, f.ObjectKey)
	if err != nil {
		_ = s.markFailed(ctx, *f)
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			return File{}, apperr.UploadExpired
		}
		return File{}, err
	}

	used, quota, err := s.quota.GetStorage(ctx, ownerID)
	if err != nil {
		return File{}, err
	}
	if used+stat.Size > quota {
		_ = s.markFailed(ctx, *f)
		return File{}, apperr.QuotaExceeded
	}

	// Replace flow: archive the old file, update the target, drop the PENDING row atomically.
	if f.ReplacesFileID != nil {
		rec, err := s.repo.CompleteReplaceUpload(ctx, ownerID, fileID, *f.ReplacesFileID, stat.Size, stat.ContentType, now)
		if err != nil {
			if errors.Is(err, apperr.QuotaExceeded) {
				_ = s.markFailed(ctx, *f)
			}
			return File{}, err
		}
		return *rec, nil
	}

	rec, err := s.repo.CompleteUpload(ctx, ownerID, fileID, stat.Size, stat.ContentType, now)
	if err != nil {
		if errors.Is(err, apperr.QuotaExceeded) {
			_ = s.markFailed(ctx, *f)
		}
		return File{}, err
	}
	s.activity.Record(ctx, ownerID, activity.TypeFileUploaded, rec.Name)
	return *rec, nil
}

func (s *Service) Get(ctx context.Context, ownerID, fileID string) (File, error) {
	f, err := s.repo.GetByID(ctx, ownerID, fileID)
	if err != nil {
		return File{}, err
	}
	if f == nil || f.DeletedAt != nil {
		return File{}, apperr.NotFound
	}
	return *f, nil
}

func (s *Service) DownloadURL(ctx context.Context, ownerID, fileID string) (DownloadURL, error) {
	f, err := s.Get(ctx, ownerID, fileID)
	if err != nil {
		return DownloadURL{}, err
	}
	if f.Status != StatusReady {
		return DownloadURL{}, apperr.NotFound
	}
	presigned, err := s.objects.CreateDownloadURL(ctx, f.ObjectKey, objectstore.DownloadOptions{
		Expires: config.DownloadPresignTTL,
	})
	if err != nil {
		return DownloadURL{}, err
	}
	return DownloadURL{URL: presigned.URL, ExpiresAt: presigned.ExpiresAt}, nil
}

// ListVersions returns archived versions of a file, newest first.
func (s *Service) ListVersions(ctx context.Context, ownerID, fileID string) ([]FileVersion, error) {
	f, err := s.Get(ctx, ownerID, fileID)
	if err != nil {
		return nil, err
	}
	if f.Status != StatusReady {
		return nil, apperr.NotFound
	}
	return s.repo.ListVersions(ctx, fileID)
}

// DownloadVersionURL returns a presigned URL for an archived version.
func (s *Service) DownloadVersionURL(ctx context.Context, ownerID, fileID, versionID string) (DownloadURL, error) {
	if _, err := s.Get(ctx, ownerID, fileID); err != nil {
		return DownloadURL{}, err
	}
	v, err := s.repo.GetVersion(ctx, fileID, versionID)
	if err != nil {
		return DownloadURL{}, err
	}
	if v == nil {
		return DownloadURL{}, apperr.NotFound
	}
	presigned, err := s.objects.CreateDownloadURL(ctx, v.ObjectKey, objectstore.DownloadOptions{
		Expires: config.DownloadPresignTTL,
	})
	if err != nil {
		return DownloadURL{}, err
	}
	return DownloadURL{URL: presigned.URL, ExpiresAt: presigned.ExpiresAt}, nil
}

func (s *Service) AbortPending(ctx context.Context, ownerID, fileID string) error {
	f, err := s.repo.GetByID(ctx, ownerID, fileID)
	if err != nil {
		return err
	}
	if f == nil || f.DeletedAt != nil {
		return apperr.NotFound
	}
	if f.Status != StatusPending {
		return apperr.InvalidState
	}
	_ = s.objects.Delete(ctx, f.ObjectKey)
	return s.repo.DeleteRow(ctx, ownerID, fileID)
}

func (s *Service) getPending(ctx context.Context, ownerID, fileID string) (*File, error) {
	f, err := s.repo.GetByID(ctx, ownerID, fileID)
	if err != nil {
		return nil, err
	}
	if f == nil || f.DeletedAt != nil {
		return nil, apperr.NotFound
	}
	if f.Status != StatusPending {
		return nil, apperr.InvalidState
	}
	return f, nil
}

func (s *Service) markFailed(ctx context.Context, f File) error {
	now := s.now().UTC()
	f.Status = StatusFailed
	f.UpdatedAt = now
	return s.repo.Update(ctx, f)
}

// SetFavorite marks an alive file as favorite; idempotent.
func (s *Service) SetFavorite(ctx context.Context, ownerID, fileID string) error {
	f, err := s.getAliveFile(ctx, ownerID, fileID)
	if err != nil {
		return err
	}
	return s.repo.AddFavorite(ctx, ownerID, f.ID, s.now().UTC())
}

// UnsetFavorite removes the favorite mark; idempotent.
func (s *Service) UnsetFavorite(ctx context.Context, ownerID, fileID string) error {
	f, err := s.getAliveFile(ctx, ownerID, fileID)
	if err != nil {
		return err
	}
	return s.repo.RemoveFavorite(ctx, ownerID, f.ID)
}

// ListFavorites returns favorited files, newest favorite first.
func (s *Service) ListFavorites(ctx context.Context, ownerID string, limit int) ([]FavoriteFile, error) {
	if limit <= 0 {
		limit = DefaultFavoritesLimit
	}
	if limit > MaxFavoritesLimit {
		limit = MaxFavoritesLimit
	}
	return s.repo.ListFavorites(ctx, ownerID, limit)
}

func (s *Service) getAliveFile(ctx context.Context, ownerID, fileID string) (*File, error) {
	f, err := s.repo.GetByID(ctx, ownerID, fileID)
	if err != nil {
		return nil, err
	}
	if f == nil || f.DeletedAt != nil || f.Status != StatusReady {
		return nil, apperr.NotFound
	}
	return f, nil
}
