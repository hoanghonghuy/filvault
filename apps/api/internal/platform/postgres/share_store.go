package postgres

import (
	"context"
	"errors"

	"filvault/internal/apperr"
	"filvault/internal/file"
	"filvault/internal/folder"
	"filvault/internal/share"
	"filvault/internal/user"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateShare(ctx context.Context, sh share.Share) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO shares (id, owner_id, resource_type, resource_id, recipient_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, sh.ID, sh.OwnerID, sh.ResourceType, sh.ResourceID, sh.RecipientID, sh.CreatedAt)
	if isUniqueViolation(err) {
		return apperr.Conflict
	}
	return err
}

func (s *Store) ListSharesByOwner(ctx context.Context, ownerID string) ([]share.Share, error) {
	rows, err := s.pool.Query(ctx, shareSelect+` WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanShares(rows)
}

func (s *Store) ListSharesByRecipient(ctx context.Context, recipientID string) ([]share.Share, error) {
	rows, err := s.pool.Query(ctx, shareSelect+` WHERE recipient_id = $1 ORDER BY created_at DESC`, recipientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanShares(rows)
}

func (s *Store) DeleteShare(ctx context.Context, ownerID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM shares WHERE id = $1 AND owner_id = $2
	`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) GetShareByRecipientResource(ctx context.Context, recipientID, resourceType, resourceID string) (*share.Share, error) {
	return scanShare(s.pool.QueryRow(ctx, shareSelect+`
		WHERE recipient_id = $1 AND resource_type = $2 AND resource_id = $3
	`, recipientID, resourceType, resourceID))
}

func (s *Store) GetShareByOwnerID(ctx context.Context, ownerID, id string) (*share.Share, error) {
	return scanShare(s.pool.QueryRow(ctx, shareSelect+`
		WHERE owner_id = $1 AND id = $2
	`, ownerID, id))
}

func (s *Store) GetFileByIDAny(ctx context.Context, id string) (*file.File, error) {
	return scanFile(s.pool.QueryRow(ctx, fileSelectAny+` AND id = $1`, id))
}

const shareSelect = `
	SELECT id, owner_id, resource_type, resource_id, recipient_id, created_at
	FROM shares
`

func scanShare(row pgx.Row) (*share.Share, error) {
	var sh share.Share
	err := row.Scan(&sh.ID, &sh.OwnerID, &sh.ResourceType, &sh.ResourceID, &sh.RecipientID, &sh.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sh, nil
}

func scanShares(rows pgx.Rows) ([]share.Share, error) {
	var list []share.Share
	for rows.Next() {
		var sh share.Share
		if err := rows.Scan(&sh.ID, &sh.OwnerID, &sh.ResourceType, &sh.ResourceID, &sh.RecipientID, &sh.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, sh)
	}
	return list, rows.Err()
}

type shareRepo struct {
	store *Store
}

func NewShareRepository(store *Store) share.Repository {
	return shareRepo{store: store}
}

func (r shareRepo) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return r.store.GetUserByEmail(ctx, email)
}

func (r shareRepo) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	return r.store.GetUserByID(ctx, id)
}

func (r shareRepo) GetAliveFileByID(ctx context.Context, ownerID, id string) (*file.File, error) {
	return r.store.GetFileByID(ctx, ownerID, id)
}

func (r shareRepo) GetAliveFolderByID(ctx context.Context, ownerID, id string) (*folder.Folder, error) {
	return r.store.GetAliveFolderByID(ctx, ownerID, id)
}

func (r shareRepo) GetFileByIDAny(ctx context.Context, id string) (*file.File, error) {
	return r.store.GetFileByIDAny(ctx, id)
}

func (r shareRepo) CreateShare(ctx context.Context, sh share.Share) error {
	return r.store.CreateShare(ctx, sh)
}

func (r shareRepo) ListSharesByOwner(ctx context.Context, ownerID string) ([]share.Share, error) {
	return r.store.ListSharesByOwner(ctx, ownerID)
}

func (r shareRepo) ListSharesByRecipient(ctx context.Context, recipientID string) ([]share.Share, error) {
	return r.store.ListSharesByRecipient(ctx, recipientID)
}

func (r shareRepo) DeleteShare(ctx context.Context, ownerID, id string) error {
	return r.store.DeleteShare(ctx, ownerID, id)
}

func (r shareRepo) GetShareByRecipientResource(ctx context.Context, recipientID, resourceType, resourceID string) (*share.Share, error) {
	return r.store.GetShareByRecipientResource(ctx, recipientID, resourceType, resourceID)
}

func (r shareRepo) GetShareByOwnerID(ctx context.Context, ownerID, id string) (*share.Share, error) {
	return r.store.GetShareByOwnerID(ctx, ownerID, id)
}

func (r shareRepo) ListAliveFolderChildren(ctx context.Context, ownerID string, parentID *string) ([]folder.Folder, error) {
	return r.store.ListAliveFolderChildren(ctx, ownerID, parentID)
}

func (r shareRepo) ListAliveFilesInFolder(ctx context.Context, ownerID string, folderID *string) ([]folder.BrowserFile, error) {
	return r.store.ListAliveFilesInFolder(ctx, ownerID, folderID)
}

var _ share.Repository = shareRepo{}
