package postgres

import (
	"context"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/file"
	"filvault/internal/folder"
	"filvault/internal/trash"
	"filvault/internal/user"

	"github.com/jackc/pgx/v5"
)

const fileSelectAny = `
	SELECT id, owner_id, folder_id, name, original_name, object_key,
		mime_type, size_bytes, status, created_at, updated_at, deleted_at, upload_expires_at, replaces_file_id
	FROM files
	WHERE 1=1
`

const folderSelectAny = `
	SELECT id, owner_id, parent_id, name, created_at, updated_at, deleted_at
	FROM folders
	WHERE 1=1
`

func (s *Store) GetAnyFileByID(ctx context.Context, ownerID, id string) (*file.File, error) {
	return scanFile(s.pool.QueryRow(ctx, fileSelectAny+` AND owner_id = $1 AND id = $2`, ownerID, id))
}

func (s *Store) SoftDeleteFile(ctx context.Context, ownerID, id string, at time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE files
		SET deleted_at = $3, updated_at = $3
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL AND status = 'READY'
	`, id, ownerID, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) RestoreFile(ctx context.Context, ownerID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE files SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NOT NULL
	`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) ListTrashedFiles(ctx context.Context, ownerID string) ([]file.File, error) {
	rows, err := s.pool.Query(ctx, fileSelectAny+`
		AND owner_id = $1 AND deleted_at IS NOT NULL
		ORDER BY deleted_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFiles(rows)
}

func (s *Store) ListExpiredTrashedFiles(ctx context.Context, ownerID string, deletedBefore time.Time) ([]file.File, error) {
	rows, err := s.pool.Query(ctx, fileSelectAny+`
		AND owner_id = $1 AND deleted_at IS NOT NULL AND deleted_at <= $2
	`, ownerID, deletedBefore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFiles(rows)
}

func scanFiles(rows pgx.Rows) ([]file.File, error) {
	var list []file.File
	for rows.Next() {
		var f file.File
		if err := rows.Scan(
			&f.ID, &f.OwnerID, &f.FolderID, &f.Name, &f.OriginalName, &f.ObjectKey,
			&f.MimeType, &f.SizeBytes, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.DeletedAt, &f.UploadExpiresAt, &f.ReplacesFileID,
		); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

func (s *Store) GetAnyFolderByID(ctx context.Context, ownerID, id string) (*folder.Folder, error) {
	return s.scanFolder(s.pool.QueryRow(ctx, folderSelectAny+` AND owner_id = $1 AND id = $2`, ownerID, id))
}

func (s *Store) RestoreFolder(ctx context.Context, ownerID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE folders SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NOT NULL
	`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) DeleteFolderRow(ctx context.Context, ownerID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM folders WHERE id = $1 AND owner_id = $2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) ListTrashedFolders(ctx context.Context, ownerID string) ([]folder.Folder, error) {
	rows, err := s.pool.Query(ctx, folderSelectAny+`
		AND owner_id = $1 AND deleted_at IS NOT NULL
		ORDER BY deleted_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFolders(rows)
}

func (s *Store) ListExpiredTrashedFolders(ctx context.Context, ownerID string, deletedBefore time.Time) ([]folder.Folder, error) {
	rows, err := s.pool.Query(ctx, folderSelectAny+`
		AND owner_id = $1 AND deleted_at IS NOT NULL AND deleted_at <= $2
	`, ownerID, deletedBefore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFolders(rows)
}

func (s *Store) CountAllFolderChildren(ctx context.Context, ownerID, folderID string) (subfolders int, files int, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM folders WHERE owner_id = $1 AND parent_id = $2
	`, ownerID, folderID).Scan(&subfolders)
	if err != nil {
		return 0, 0, err
	}
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM files WHERE owner_id = $1 AND folder_id = $2
	`, ownerID, folderID).Scan(&files)
	return subfolders, files, err
}

func (s *Store) ExistsAliveFolderByName(ctx context.Context, ownerID string, parentID *string, name, excludeFolderID string) (bool, error) {
	var exists bool
	var err error
	if parentID == nil {
		err = s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM folders
				WHERE owner_id = $1 AND parent_id IS NULL AND name = $2
					AND deleted_at IS NULL AND ($3 = '' OR id <> $3)
			)
		`, ownerID, name, excludeFolderID).Scan(&exists)
	} else {
		err = s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM folders
				WHERE owner_id = $1 AND parent_id = $2 AND name = $3
					AND deleted_at IS NULL AND ($4 = '' OR id <> $4)
			)
		`, ownerID, *parentID, name, excludeFolderID).Scan(&exists)
	}
	return exists, err
}

func (s *Store) ListUsersWithAutoTrash(ctx context.Context) ([]user.User, error) {
	rows, err := s.pool.Query(ctx, userSelect+` WHERE trash_auto_delete_enabled = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []user.User
	for rows.Next() {
		u, err := s.scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *u)
	}
	return list, rows.Err()
}

type trashRepo struct {
	store *Store
}

func NewTrashRepository(store *Store) trash.Repository {
	return trashRepo{store: store}
}

func (r trashRepo) GetAnyFile(ctx context.Context, ownerID, id string) (*file.File, error) {
	return r.store.GetAnyFileByID(ctx, ownerID, id)
}

func (r trashRepo) SoftDeleteFile(ctx context.Context, ownerID, id string, at time.Time) error {
	return r.store.SoftDeleteFile(ctx, ownerID, id, at)
}

func (r trashRepo) RestoreFile(ctx context.Context, ownerID, id string) error {
	return r.store.RestoreFile(ctx, ownerID, id)
}

func (r trashRepo) DeleteFileRow(ctx context.Context, ownerID, id string) error {
	return r.store.DeleteFileRow(ctx, ownerID, id)
}

func (r trashRepo) ListTrashedFiles(ctx context.Context, ownerID string) ([]file.File, error) {
	return r.store.ListTrashedFiles(ctx, ownerID)
}

func (r trashRepo) ListExpiredTrashedFiles(ctx context.Context, ownerID string, deletedBefore time.Time) ([]file.File, error) {
	return r.store.ListExpiredTrashedFiles(ctx, ownerID, deletedBefore)
}

func (r trashRepo) ExistsAliveFileByName(ctx context.Context, ownerID string, folderID *string, name, excludeFileID string) (bool, error) {
	return r.store.ExistsAliveFileByName(ctx, ownerID, folderID, name, excludeFileID)
}

func (r trashRepo) GetAnyFolder(ctx context.Context, ownerID, id string) (*folder.Folder, error) {
	return r.store.GetAnyFolderByID(ctx, ownerID, id)
}

func (r trashRepo) GetAliveFolderByID(ctx context.Context, ownerID, id string) (*folder.Folder, error) {
	return r.store.GetAliveFolderByID(ctx, ownerID, id)
}

func (r trashRepo) RestoreFolder(ctx context.Context, ownerID, id string) error {
	return r.store.RestoreFolder(ctx, ownerID, id)
}

func (r trashRepo) DeleteFolderRow(ctx context.Context, ownerID, id string) error {
	return r.store.DeleteFolderRow(ctx, ownerID, id)
}

func (r trashRepo) ListTrashedFolders(ctx context.Context, ownerID string) ([]folder.Folder, error) {
	return r.store.ListTrashedFolders(ctx, ownerID)
}

func (r trashRepo) ListExpiredTrashedFolders(ctx context.Context, ownerID string, deletedBefore time.Time) ([]folder.Folder, error) {
	return r.store.ListExpiredTrashedFolders(ctx, ownerID, deletedBefore)
}

func (r trashRepo) CountAllFolderChildren(ctx context.Context, ownerID, folderID string) (int, int, error) {
	return r.store.CountAllFolderChildren(ctx, ownerID, folderID)
}

func (r trashRepo) ExistsAliveFolderByName(ctx context.Context, ownerID string, parentID *string, name, excludeFolderID string) (bool, error) {
	return r.store.ExistsAliveFolderByName(ctx, ownerID, parentID, name, excludeFolderID)
}

func (r trashRepo) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	return r.store.GetUserByID(ctx, id)
}

func (r trashRepo) ListUsersWithAutoTrash(ctx context.Context) ([]user.User, error) {
	return r.store.ListUsersWithAutoTrash(ctx)
}

func (r trashRepo) UpdateTrashSettings(ctx context.Context, userID string, enabled bool, retentionDays int) error {
	return r.store.UpdateTrashSettings(ctx, userID, enabled, retentionDays)
}

var _ trash.Repository = trashRepo{}
