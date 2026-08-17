package file

import (
	"context"
	"errors"
	"strings"
	"time"

	"filnest/internal/apperr"
	"filnest/internal/auth"
	"filnest/internal/folder"
	"filnest/internal/platform/config"
	"filnest/internal/platform/objectstore"
)

type Service struct {
	repo    Repository
	folders folder.Repository
	quota   QuotaStore
	objects objectstore.ObjectStore
	now     func() time.Time
}

func NewService(repo Repository, folders folder.Repository, quota QuotaStore, objects objectstore.ObjectStore) *Service {
	return &Service{
		repo:    repo,
		folders: folders,
		quota:   quota,
		objects: objects,
		now:     time.Now,
	}
}

func (s *Service) CreateUploadSession(ctx context.Context, ownerID, name, contentType string, size int64, folderID *string) (UploadSession, error) {
	if err := ValidateUpload(name, contentType, size); err != nil {
		return UploadSession{}, err
	}
	if folderID != nil && *folderID == "" {
		folderID = nil
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
	exists, err := s.repo.ExistsAliveByName(ctx, ownerID, folderID, displayName, "")
	if err != nil {
		return UploadSession{}, err
	}
	if exists {
		return UploadSession{}, apperr.Conflict
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

	rec := *f
	rec.Status = StatusReady
	rec.SizeBytes = stat.Size
	if stat.ContentType != "" {
		rec.MimeType = stat.ContentType
	}
	rec.UploadExpiresAt = nil
	rec.UpdatedAt = now
	if err := s.repo.Update(ctx, rec); err != nil {
		return File{}, err
	}
	if err := s.quota.AddStorageUsed(ctx, ownerID, stat.Size); err != nil {
		return File{}, err
	}
	return rec, nil
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
