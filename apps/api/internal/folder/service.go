package folder

import (
	"context"
	"strings"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) Create(ctx context.Context, ownerID, name string, parentID *string) (Folder, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Folder{}, apperr.Validation
	}
	if parentID != nil && *parentID == "" {
		parentID = nil
	}
	if err := s.validateParent(ctx, ownerID, parentID); err != nil {
		return Folder{}, err
	}
	now := s.now().UTC()
	f := Folder{
		ID:        auth.NewID(),
		OwnerID:   ownerID,
		ParentID:  parentID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return Folder{}, err
	}
	return f, nil
}

func (s *Service) Get(ctx context.Context, ownerID, id string) (Folder, error) {
	f, err := s.repo.GetAliveByID(ctx, ownerID, id)
	if err != nil {
		return Folder{}, err
	}
	if f == nil {
		return Folder{}, apperr.NotFound
	}
	return *f, nil
}

func (s *Service) Patch(ctx context.Context, ownerID, id string, name *string, moveParent bool, newParentID *string) (Folder, error) {
	f, err := s.Get(ctx, ownerID, id)
	if err != nil {
		return Folder{}, err
	}
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return Folder{}, apperr.Validation
		}
		f.Name = trimmed
	}
	if moveParent {
		parent := newParentID
		if parent != nil && *parent == "" {
			parent = nil
		}
		if parent != nil && *parent == id {
			return Folder{}, apperr.Validation
		}
		if err := s.validateParent(ctx, ownerID, parent); err != nil {
			return Folder{}, err
		}
		if parent != nil {
			isDesc, err := s.repo.IsAncestor(ctx, ownerID, id, *parent)
			if err != nil {
				return Folder{}, err
			}
			if isDesc {
				return Folder{}, apperr.Validation
			}
		}
		f.ParentID = parent
	}
	f.UpdatedAt = s.now().UTC()
	if err := s.repo.Update(ctx, f); err != nil {
		return Folder{}, err
	}
	return f, nil
}

func (s *Service) Delete(ctx context.Context, ownerID, id string) error {
	f, err := s.Get(ctx, ownerID, id)
	if err != nil {
		return err
	}
	subfolders, files, err := s.repo.CountAliveChildren(ctx, ownerID, f.ID)
	if err != nil {
		return err
	}
	if subfolders > 0 || files > 0 {
		return apperr.Conflict
	}
	return s.repo.SoftDelete(ctx, ownerID, id, s.now().UTC())
}

func (s *Service) Browser(ctx context.Context, ownerID string, folderID *string) (Browser, error) {
	if folderID != nil && *folderID == "" {
		folderID = nil
	}
	var result Browser
	listParent := folderID
	if folderID != nil {
		f, err := s.Get(ctx, ownerID, *folderID)
		if err != nil {
			return Browser{}, err
		}
		result.Folder = &f
		crumb, err := s.breadcrumb(ctx, ownerID, f)
		if err != nil {
			return Browser{}, err
		}
		result.Breadcrumb = crumb
		listParent = &f.ID
	}
	folders, err := s.repo.ListAliveChildren(ctx, ownerID, listParent)
	if err != nil {
		return Browser{}, err
	}
	files, err := s.repo.ListAliveFilesInFolder(ctx, ownerID, listParent)
	if err != nil {
		return Browser{}, err
	}
	result.Folders = folders
	result.Files = files
	return result, nil
}

func (s *Service) breadcrumb(ctx context.Context, ownerID string, f Folder) ([]Folder, error) {
	var chain []Folder
	current := f
	for current.ParentID != nil {
		parent, err := s.repo.GetAliveByID(ctx, ownerID, *current.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			break
		}
		chain = append([]Folder{*parent}, chain...)
		current = *parent
	}
	return chain, nil
}

func (s *Service) validateParent(ctx context.Context, ownerID string, parentID *string) error {
	if parentID == nil {
		return nil
	}
	parent, err := s.repo.GetAliveByID(ctx, ownerID, *parentID)
	if err != nil {
		return err
	}
	if parent == nil {
		return apperr.NotFound
	}
	return nil
}
