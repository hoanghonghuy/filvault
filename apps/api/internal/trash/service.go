package trash

import (
	"context"
	"time"

	"filvault/internal/activity"
	"filvault/internal/apperr"
	"filvault/internal/file"
	"filvault/internal/folder"
)

type Service struct {
	repo     Repository
	quota    QuotaStore
	objects  ObjectStore
	activity ActivityRecorder
	now      func() time.Time
}

// ActivityRecorder records lifecycle events; failures never break trash ops.
type ActivityRecorder interface {
	Record(ctx context.Context, ownerID, eventType, targetName string)
}

func NewService(repo Repository, quota QuotaStore, objects ObjectStore, activity ActivityRecorder) *Service {
	return &Service{repo: repo, quota: quota, objects: objects, activity: activity, now: time.Now}
}

func (s *Service) List(ctx context.Context, ownerID string) (List, error) {
	folders, err := s.repo.ListTrashedFolders(ctx, ownerID)
	if err != nil {
		return List{}, err
	}
	files, err := s.repo.ListTrashedFiles(ctx, ownerID)
	if err != nil {
		return List{}, err
	}
	out := List{
		Folders: make([]Item, 0, len(folders)),
		Files:   make([]Item, 0, len(files)),
	}
	for _, f := range folders {
		out.Folders = append(out.Folders, folderItem(f))
	}
	for _, f := range files {
		out.Files = append(out.Files, fileItem(f))
	}
	return out, nil
}

func (s *Service) RestoreFile(ctx context.Context, ownerID, fileID string) error {
	f, err := s.getTrashedFile(ctx, ownerID, fileID)
	if err != nil {
		return err
	}
	if f.FolderID != nil {
		parent, err := s.repo.GetAliveFolderByID(ctx, ownerID, *f.FolderID)
		if err != nil {
			return err
		}
		if parent == nil {
			return apperr.NotFound
		}
	}
	exists, err := s.repo.ExistsAliveFileByName(ctx, ownerID, f.FolderID, f.Name, f.ID)
	if err != nil {
		return err
	}
	if exists {
		return apperr.Conflict
	}
	if err := s.repo.RestoreFile(ctx, ownerID, fileID); err != nil {
		return err
	}
	s.activity.Record(ctx, ownerID, activity.TypeFileRestored, f.Name)
	return nil
}

func (s *Service) RestoreFolder(ctx context.Context, ownerID, folderID string) error {
	f, err := s.getTrashedFolder(ctx, ownerID, folderID)
	if err != nil {
		return err
	}
	if f.ParentID != nil {
		parent, err := s.repo.GetAliveFolderByID(ctx, ownerID, *f.ParentID)
		if err != nil {
			return err
		}
		if parent == nil {
			return apperr.NotFound
		}
	}
	exists, err := s.repo.ExistsAliveFolderByName(ctx, ownerID, f.ParentID, f.Name, f.ID)
	if err != nil {
		return err
	}
	if exists {
		return apperr.Conflict
	}
	s.activity.Record(ctx, ownerID, activity.TypeFolderRestored, f.Name)
	return s.repo.RestoreFolder(ctx, ownerID, folderID)
}

func (s *Service) PermanentDeleteFile(ctx context.Context, ownerID, fileID string) error {
	f, err := s.getTrashedFile(ctx, ownerID, fileID)
	if err != nil && err != apperr.NotFound {
		return err
	}
	_, err = s.repo.PurgeFile(ctx, ownerID, fileID)
	if err != nil {
		return err
	}
	if f != nil {
		s.activity.Record(ctx, ownerID, activity.TypePurged, f.Name)
	}
	return nil
}

func (s *Service) PermanentDeleteFolder(ctx context.Context, ownerID, folderID string) error {
	if _, err := s.getTrashedFolder(ctx, ownerID, folderID); err != nil {
		return err
	}
	sub, files, err := s.repo.CountAllFolderChildren(ctx, ownerID, folderID)
	if err != nil {
		return err
	}
	if sub > 0 || files > 0 {
		return apperr.Conflict
	}
	return s.repo.DeleteFolderRow(ctx, ownerID, folderID)
}

func (s *Service) RunAutoCleanup(ctx context.Context) error {
	users, err := s.repo.ListUsersWithAutoTrash(ctx)
	if err != nil {
		return err
	}
	now := s.now().UTC()
	for _, u := range users {
		if u.TrashRetentionDays < 1 {
			continue
		}
		cutoff := now.Add(-time.Duration(u.TrashRetentionDays) * 24 * time.Hour)
		files, err := s.repo.ListExpiredTrashedFiles(ctx, u.ID, cutoff)
		if err != nil {
			return err
		}
		for _, f := range files {
			if err := s.PermanentDeleteFile(ctx, u.ID, f.ID); err != nil {
				return err
			}
		}
		folders, err := s.repo.ListExpiredTrashedFolders(ctx, u.ID, cutoff)
		if err != nil {
			return err
		}
		for _, f := range folders {
			sub, childFiles, err := s.repo.CountAllFolderChildren(ctx, u.ID, f.ID)
			if err != nil {
				return err
			}
			if sub > 0 || childFiles > 0 {
				continue
			}
			if err := s.repo.DeleteFolderRow(ctx, u.ID, f.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) getTrashedFile(ctx context.Context, ownerID, id string) (*file.File, error) {
	f, err := s.repo.GetAnyFile(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if f == nil || f.DeletedAt == nil {
		return nil, apperr.NotFound
	}
	return f, nil
}

func (s *Service) getTrashedFolder(ctx context.Context, ownerID, id string) (*folder.Folder, error) {
	f, err := s.repo.GetAnyFolder(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if f == nil || f.DeletedAt == nil {
		return nil, apperr.NotFound
	}
	return f, nil
}
