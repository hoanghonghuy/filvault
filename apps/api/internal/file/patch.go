package file

import (
	"context"
	"strings"

	"filnest/internal/apperr"
)

func (s *Service) Patch(ctx context.Context, ownerID, fileID string, name *string, moveFolder bool, newFolderID *string) (File, error) {
	f, err := s.getReady(ctx, ownerID, fileID)
	if err != nil {
		return File{}, err
	}
	rec := *f
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return File{}, apperr.Validation
		}
		rec.Name = trimmed
	}
	if moveFolder {
		folderID := newFolderID
		if folderID != nil && *folderID == "" {
			folderID = nil
		}
		if err := s.validateFolder(ctx, ownerID, folderID); err != nil {
			return File{}, err
		}
		rec.FolderID = folderID
	}
	exists, err := s.repo.ExistsAliveByName(ctx, ownerID, rec.FolderID, rec.Name, rec.ID)
	if err != nil {
		return File{}, err
	}
	if exists {
		return File{}, apperr.Conflict
	}
	rec.UpdatedAt = s.now().UTC()
	if err := s.repo.Update(ctx, rec); err != nil {
		return File{}, err
	}
	return rec, nil
}

func (s *Service) getReady(ctx context.Context, ownerID, fileID string) (*File, error) {
	f, err := s.repo.GetByID(ctx, ownerID, fileID)
	if err != nil {
		return nil, err
	}
	if f == nil || f.DeletedAt != nil {
		return nil, apperr.NotFound
	}
	if f.Status != StatusReady {
		return nil, apperr.InvalidState
	}
	return f, nil
}

func (s *Service) validateFolder(ctx context.Context, ownerID string, folderID *string) error {
	if folderID == nil {
		return nil
	}
	parent, err := s.folders.GetAliveByID(ctx, ownerID, *folderID)
	if err != nil {
		return err
	}
	if parent == nil {
		return apperr.NotFound
	}
	return nil
}
