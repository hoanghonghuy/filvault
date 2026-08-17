package postgres

import (
	"context"
	"errors"
	"time"

	"filnest/internal/apperr"
	"filnest/internal/folder"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateFolder(ctx context.Context, f folder.Folder) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO folders (id, owner_id, parent_id, name, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULL)
	`, f.ID, f.OwnerID, f.ParentID, f.Name, f.CreatedAt, f.UpdatedAt)
	if isUniqueViolation(err) {
		return apperr.Conflict
	}
	return err
}

func (s *Store) GetAliveFolderByID(ctx context.Context, ownerID, id string) (*folder.Folder, error) {
	return s.scanFolder(s.pool.QueryRow(ctx, folderSelectAlive+` AND owner_id = $1 AND id = $2`, ownerID, id))
}

func (s *Store) UpdateFolder(ctx context.Context, f folder.Folder) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE folders
		SET name = $2, parent_id = $3, updated_at = $4
		WHERE id = $1 AND owner_id = $5 AND deleted_at IS NULL
	`, f.ID, f.Name, f.ParentID, f.UpdatedAt, f.OwnerID)
	if isUniqueViolation(err) {
		return apperr.Conflict
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) SoftDeleteFolder(ctx context.Context, ownerID, id string, at time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE folders
		SET deleted_at = $3, updated_at = $3
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL
	`, id, ownerID, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) ListAliveFolderChildren(ctx context.Context, ownerID string, parentID *string) ([]folder.Folder, error) {
	var rows pgx.Rows
	var err error
	if parentID == nil {
		rows, err = s.pool.Query(ctx, folderSelectAlive+` AND owner_id = $1 AND parent_id IS NULL ORDER BY name`, ownerID)
	} else {
		rows, err = s.pool.Query(ctx, folderSelectAlive+` AND owner_id = $1 AND parent_id = $2 ORDER BY name`, ownerID, *parentID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFolders(rows)
}

func (s *Store) CountAliveFolderChildren(ctx context.Context, ownerID, folderID string) (subfolders int, files int, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM folders
		WHERE owner_id = $1 AND parent_id = $2 AND deleted_at IS NULL
	`, ownerID, folderID).Scan(&subfolders)
	if err != nil {
		return 0, 0, err
	}
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM files
		WHERE owner_id = $1 AND folder_id = $2 AND deleted_at IS NULL AND status = 'READY'
	`, ownerID, folderID).Scan(&files)
	return subfolders, files, err
}

func (s *Store) ListAliveFilesInFolder(ctx context.Context, ownerID string, folderID *string) ([]folder.BrowserFile, error) {
	var rows pgx.Rows
	var err error
	if folderID == nil {
		rows, err = s.pool.Query(ctx, `
			SELECT id, name, mime_type, size_bytes, updated_at
			FROM files
			WHERE owner_id = $1 AND folder_id IS NULL AND deleted_at IS NULL AND status = 'READY'
			ORDER BY name
		`, ownerID)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id, name, mime_type, size_bytes, updated_at
			FROM files
			WHERE owner_id = $1 AND folder_id = $2 AND deleted_at IS NULL AND status = 'READY'
			ORDER BY name
		`, ownerID, *folderID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []folder.BrowserFile
	for rows.Next() {
		var f folder.BrowserFile
		if err := rows.Scan(&f.ID, &f.Name, &f.MimeType, &f.SizeBytes, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Store) IsFolderAncestor(ctx context.Context, ownerID, ancestorID, nodeID string) (bool, error) {
	current := nodeID
	for current != "" {
		if current == ancestorID {
			return true, nil
		}
		f, err := s.GetAliveFolderByID(ctx, ownerID, current)
		if err != nil {
			return false, err
		}
		if f == nil || f.ParentID == nil {
			return false, nil
		}
		current = *f.ParentID
	}
	return false, nil
}

const folderSelectAlive = `
	SELECT id, owner_id, parent_id, name, created_at, updated_at, deleted_at
	FROM folders
	WHERE deleted_at IS NULL
`

func (s *Store) scanFolder(row pgx.Row) (*folder.Folder, error) {
	var f folder.Folder
	err := row.Scan(&f.ID, &f.OwnerID, &f.ParentID, &f.Name, &f.CreatedAt, &f.UpdatedAt, &f.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func scanFolders(rows pgx.Rows) ([]folder.Folder, error) {
	var list []folder.Folder
	for rows.Next() {
		var f folder.Folder
		if err := rows.Scan(&f.ID, &f.OwnerID, &f.ParentID, &f.Name, &f.CreatedAt, &f.UpdatedAt, &f.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

// folderRepo adapts Store to folder.Repository.
type folderRepo struct {
	store *Store
}

func (r folderRepo) Create(ctx context.Context, f folder.Folder) error {
	return r.store.CreateFolder(ctx, f)
}

func (r folderRepo) GetAliveByID(ctx context.Context, ownerID, id string) (*folder.Folder, error) {
	return r.store.GetAliveFolderByID(ctx, ownerID, id)
}

func (r folderRepo) Update(ctx context.Context, f folder.Folder) error {
	return r.store.UpdateFolder(ctx, f)
}

func (r folderRepo) SoftDelete(ctx context.Context, ownerID, id string, at time.Time) error {
	return r.store.SoftDeleteFolder(ctx, ownerID, id, at)
}

func (r folderRepo) ListAliveChildren(ctx context.Context, ownerID string, parentID *string) ([]folder.Folder, error) {
	return r.store.ListAliveFolderChildren(ctx, ownerID, parentID)
}

func (r folderRepo) CountAliveChildren(ctx context.Context, ownerID, folderID string) (int, int, error) {
	return r.store.CountAliveFolderChildren(ctx, ownerID, folderID)
}

func (r folderRepo) ListAliveFilesInFolder(ctx context.Context, ownerID string, folderID *string) ([]folder.BrowserFile, error) {
	return r.store.ListAliveFilesInFolder(ctx, ownerID, folderID)
}

func (r folderRepo) IsAncestor(ctx context.Context, ownerID, ancestorID, nodeID string) (bool, error) {
	return r.store.IsFolderAncestor(ctx, ownerID, ancestorID, nodeID)
}

func NewFolderRepository(store *Store) folder.Repository {
	return folderRepo{store: store}
}

var _ folder.Repository = folderRepo{}
