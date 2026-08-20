package share

import (
	"context"
	"strings"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/file"
	"filvault/internal/folder"
	"filvault/internal/platform/config"
	"filvault/internal/platform/objectstore"
	"filvault/internal/user"
)

type Service struct {
	repo    Repository
	objects ObjectStore
	now     func() time.Time
}

func NewService(repo Repository, objects ObjectStore) *Service {
	return &Service{repo: repo, objects: objects, now: time.Now}
}

// Create shares a file or folder with the user identified by email.
func (s *Service) Create(ctx context.Context, ownerID, resourceType, resourceID, email string) (Share, error) {
	resourceType = strings.ToLower(strings.TrimSpace(resourceType))
	if resourceType != ResourceFile && resourceType != ResourceFolder {
		return Share{}, apperr.Validation
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return Share{}, apperr.Validation
	}
	recipient, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return Share{}, err
	}
	if recipient == nil {
		return Share{}, apperr.NotFound
	}
	if recipient.ID == ownerID {
		return Share{}, apperr.Validation
	}
	if err := s.validateResource(ctx, ownerID, resourceType, resourceID); err != nil {
		return Share{}, err
	}
	sh := Share{
		ID:           auth.NewID(),
		OwnerID:      ownerID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		RecipientID:  recipient.ID,
		CreatedAt:    s.now().UTC(),
	}
	if err := s.repo.CreateShare(ctx, sh); err != nil {
		return Share{}, err
	}
	return sh, nil
}

// ListOutgoing returns shares the owner created, with resource names resolved.
func (s *Service) ListOutgoing(ctx context.Context, ownerID string) ([]Outgoing, error) {
	shares, err := s.repo.ListSharesByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]Outgoing, 0, len(shares))
	for _, sh := range shares {
		recipient, err := s.repo.GetUserByID(ctx, sh.RecipientID)
		if err != nil {
			return nil, err
		}
		name, err := s.resourceName(ctx, sh.OwnerID, sh.ResourceType, sh.ResourceID)
		if err != nil {
			return nil, err
		}
		out = append(out, Outgoing{
			ID:           sh.ID,
			ResourceType: sh.ResourceType,
			ResourceID:   sh.ResourceID,
			ResourceName: name,
			Recipient:    userRef(recipient),
			CreatedAt:    sh.CreatedAt.UTC(),
		})
	}
	return out, nil
}

// ListIncoming returns shares the recipient received, with owner info resolved.
func (s *Service) ListIncoming(ctx context.Context, recipientID string) ([]Incoming, error) {
	shares, err := s.repo.ListSharesByRecipient(ctx, recipientID)
	if err != nil {
		return nil, err
	}
	out := make([]Incoming, 0, len(shares))
	for _, sh := range shares {
		owner, err := s.repo.GetUserByID(ctx, sh.OwnerID)
		if err != nil {
			return nil, err
		}
		name, err := s.resourceName(ctx, sh.OwnerID, sh.ResourceType, sh.ResourceID)
		if err != nil {
			return nil, err
		}
		out = append(out, Incoming{
			ID:           sh.ID,
			ResourceType: sh.ResourceType,
			ResourceID:   sh.ResourceID,
			ResourceName: name,
			Owner:        userRef(owner),
			CreatedAt:    sh.CreatedAt.UTC(),
		})
	}
	return out, nil
}

// Revoke removes a share. Only the owner may revoke.
func (s *Service) Revoke(ctx context.Context, ownerID, id string) error {
	return s.repo.DeleteShare(ctx, ownerID, id)
}

// BrowseFolder lists the direct children of a folder shared with the recipient.
func (s *Service) BrowseFolder(ctx context.Context, recipientID, folderID string) (folder.Browser, error) {
	sh, err := s.repo.GetShareByRecipientResource(ctx, recipientID, ResourceFolder, folderID)
	if err != nil {
		return folder.Browser{}, err
	}
	if sh == nil {
		return folder.Browser{}, apperr.NotFound
	}
	f, err := s.repo.GetAliveFolderByID(ctx, sh.OwnerID, folderID)
	if err != nil {
		return folder.Browser{}, err
	}
	if f == nil {
		return folder.Browser{}, apperr.NotFound
	}
	folders, err := s.repo.ListAliveFolderChildren(ctx, sh.OwnerID, &folderID)
	if err != nil {
		return folder.Browser{}, err
	}
	files, err := s.repo.ListAliveFilesInFolder(ctx, sh.OwnerID, &folderID)
	if err != nil {
		return folder.Browser{}, err
	}
	return folder.Browser{
		Folder:  f,
		Folders: folders,
		Files:   files,
	}, nil
}

// DownloadURL returns a presigned URL for a file shared with the recipient,
// either directly or via a shared parent folder.
func (s *Service) DownloadURL(ctx context.Context, recipientID, fileID string) (DownloadURL, error) {
	f, err := s.repo.GetFileByIDAny(ctx, fileID)
	if err != nil {
		return DownloadURL{}, err
	}
	if f == nil || f.DeletedAt != nil || f.Status != file.StatusReady {
		return DownloadURL{}, apperr.NotFound
	}
	allowed, err := s.canAccessFile(ctx, recipientID, f)
	if err != nil {
		return DownloadURL{}, err
	}
	if !allowed {
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

func (s *Service) canAccessFile(ctx context.Context, recipientID string, f *file.File) (bool, error) {
	direct, err := s.repo.GetShareByRecipientResource(ctx, recipientID, ResourceFile, f.ID)
	if err != nil {
		return false, err
	}
	if direct != nil {
		return true, nil
	}
	if f.FolderID != nil {
		viaFolder, err := s.repo.GetShareByRecipientResource(ctx, recipientID, ResourceFolder, *f.FolderID)
		if err != nil {
			return false, err
		}
		if viaFolder != nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) validateResource(ctx context.Context, ownerID, resourceType, resourceID string) error {
	switch resourceType {
	case ResourceFile:
		f, err := s.repo.GetAliveFileByID(ctx, ownerID, resourceID)
		if err != nil {
			return err
		}
		if f == nil {
			return apperr.NotFound
		}
	case ResourceFolder:
		f, err := s.repo.GetAliveFolderByID(ctx, ownerID, resourceID)
		if err != nil {
			return err
		}
		if f == nil {
			return apperr.NotFound
		}
	}
	return nil
}

func (s *Service) resourceName(ctx context.Context, ownerID, resourceType, resourceID string) (string, error) {
	switch resourceType {
	case ResourceFile:
		f, err := s.repo.GetFileByIDAny(ctx, resourceID)
		if err != nil {
			return "", err
		}
		if f == nil {
			return "", nil
		}
		return f.Name, nil
	case ResourceFolder:
		f, err := s.repo.GetAliveFolderByID(ctx, ownerID, resourceID)
		if err != nil {
			return "", err
		}
		if f == nil {
			return "", nil
		}
		return f.Name, nil
	}
	return "", nil
}

func userRef(u *user.User) UserRef {
	if u == nil {
		return UserRef{}
	}
	return UserRef{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName}
}
