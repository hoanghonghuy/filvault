package postgres

import (
	"context"
	"errors"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/file"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateFile(ctx context.Context, f file.File) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO files (
			id, owner_id, folder_id, name, original_name, object_key,
			mime_type, size_bytes, status, created_at, updated_at, deleted_at, upload_expires_at, replaces_file_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NULL, $12, $13)
	`, f.ID, f.OwnerID, f.FolderID, f.Name, f.OriginalName, f.ObjectKey,
		f.MimeType, f.SizeBytes, f.Status, f.CreatedAt, f.UpdatedAt, f.UploadExpiresAt, f.ReplacesFileID)
	if isUniqueViolation(err) {
		return apperr.Conflict
	}
	return err
}

func (s *Store) GetFileByID(ctx context.Context, ownerID, id string) (*file.File, error) {
	return scanFile(s.pool.QueryRow(ctx, fileSelect+` AND owner_id = $1 AND id = $2`, ownerID, id))
}

func (s *Store) UpdateFile(ctx context.Context, f file.File) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE files
		SET name = $2, folder_id = $3, mime_type = $4, size_bytes = $5, status = $6,
			updated_at = $7, upload_expires_at = $8
		WHERE id = $1 AND owner_id = $9 AND deleted_at IS NULL
	`, f.ID, f.Name, f.FolderID, f.MimeType, f.SizeBytes, f.Status, f.UpdatedAt, f.UploadExpiresAt, f.OwnerID)
	return err
}

func (s *Store) DeleteFileRow(ctx context.Context, ownerID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM files WHERE id = $1 AND owner_id = $2
	`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) ExistsAliveFileByName(ctx context.Context, ownerID string, folderID *string, name, excludeFileID string) (bool, error) {
	var exists bool
	var err error
	if folderID == nil {
		err = s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM files
				WHERE owner_id = $1 AND folder_id IS NULL AND name = $2
					AND deleted_at IS NULL AND status IN ('PENDING', 'READY')
					AND ($3 = '' OR id <> $3)
			)
		`, ownerID, name, excludeFileID).Scan(&exists)
	} else {
		err = s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM files
				WHERE owner_id = $1 AND folder_id = $2 AND name = $3
					AND deleted_at IS NULL AND status IN ('PENDING', 'READY')
					AND ($4 = '' OR id <> $4)
			)
		`, ownerID, *folderID, name, excludeFileID).Scan(&exists)
	}
	return exists, err
}

func (s *Store) GetUserStorage(ctx context.Context, userID string) (used, quota int64, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT storage_used, storage_quota FROM users WHERE id = $1
	`, userID).Scan(&used, &quota)
	return used, quota, err
}

func (s *Store) AddUserStorageUsed(ctx context.Context, userID string, delta int64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE users SET storage_used = storage_used + $2, updated_at = now()
		WHERE id = $1
	`, userID, delta)
	return err
}

func (s *Store) CreateFileVersion(ctx context.Context, v file.FileVersion) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO file_versions (id, file_id, object_key, size_bytes, mime_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, v.ID, v.FileID, v.ObjectKey, v.SizeBytes, v.MimeType, v.CreatedAt)
	return err
}

func (s *Store) ListFileVersions(ctx context.Context, fileID string) ([]file.FileVersion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, file_id, object_key, size_bytes, mime_type, created_at
		FROM file_versions
		WHERE file_id = $1
		ORDER BY created_at DESC
	`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []file.FileVersion
	for rows.Next() {
		var v file.FileVersion
		if err := rows.Scan(&v.ID, &v.FileID, &v.ObjectKey, &v.SizeBytes, &v.MimeType, &v.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, rows.Err()
}

func (s *Store) GetFileVersion(ctx context.Context, fileID, versionID string) (*file.FileVersion, error) {
	var v file.FileVersion
	err := s.pool.QueryRow(ctx, `
		SELECT id, file_id, object_key, size_bytes, mime_type, created_at
		FROM file_versions
		WHERE file_id = $1 AND id = $2
	`, fileID, versionID).Scan(&v.ID, &v.FileID, &v.ObjectKey, &v.SizeBytes, &v.MimeType, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

const fileSelect = `
	SELECT id, owner_id, folder_id, name, original_name, object_key,
		mime_type, size_bytes, status, created_at, updated_at, deleted_at, upload_expires_at, replaces_file_id
	FROM files
	WHERE deleted_at IS NULL
`

func scanFile(row pgx.Row) (*file.File, error) {
	var f file.File
	err := row.Scan(
		&f.ID, &f.OwnerID, &f.FolderID, &f.Name, &f.OriginalName, &f.ObjectKey,
		&f.MimeType, &f.SizeBytes, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.DeletedAt, &f.UploadExpiresAt, &f.ReplacesFileID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

type fileRepo struct {
	store *Store
}

func NewFileRepository(store *Store) file.Repository {
	return fileRepo{store: store}
}

func (r fileRepo) Create(ctx context.Context, f file.File) error {
	return r.store.CreateFile(ctx, f)
}

func (r fileRepo) GetByID(ctx context.Context, ownerID, id string) (*file.File, error) {
	return r.store.GetFileByID(ctx, ownerID, id)
}

func (r fileRepo) Update(ctx context.Context, f file.File) error {
	return r.store.UpdateFile(ctx, f)
}

func (r fileRepo) DeleteRow(ctx context.Context, ownerID, id string) error {
	return r.store.DeleteFileRow(ctx, ownerID, id)
}

func (r fileRepo) SoftDelete(ctx context.Context, ownerID, id string, at time.Time) error {
	return r.store.SoftDeleteFile(ctx, ownerID, id, at)
}

func (r fileRepo) ExistsAliveByName(ctx context.Context, ownerID string, folderID *string, name, excludeFileID string) (bool, error) {
	return r.store.ExistsAliveFileByName(ctx, ownerID, folderID, name, excludeFileID)
}

func (r fileRepo) CreateVersion(ctx context.Context, v file.FileVersion) error {
	return r.store.CreateFileVersion(ctx, v)
}

func (r fileRepo) ListVersions(ctx context.Context, fileID string) ([]file.FileVersion, error) {
	return r.store.ListFileVersions(ctx, fileID)
}

func (r fileRepo) GetVersion(ctx context.Context, fileID, versionID string) (*file.FileVersion, error) {
	return r.store.GetFileVersion(ctx, fileID, versionID)
}

type quotaRepo struct {
	store *Store
}

func NewQuotaStore(store *Store) file.QuotaStore {
	return quotaRepo{store: store}
}

func (r quotaRepo) GetStorage(ctx context.Context, userID string) (int64, int64, error) {
	return r.store.GetUserStorage(ctx, userID)
}

func (r quotaRepo) AddStorageUsed(ctx context.Context, userID string, delta int64) error {
	return r.store.AddUserStorageUsed(ctx, userID, delta)
}

var _ file.Repository = fileRepo{}
var _ file.QuotaStore = quotaRepo{}
