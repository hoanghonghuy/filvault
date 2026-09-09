package postgres

import (
	"context"
	"errors"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/file"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateFile(ctx context.Context, f file.File) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO files (
			id, owner_id, folder_id, name, original_name, object_key,
			mime_type, size_bytes, status, created_at, updated_at, deleted_at, upload_expires_at, replaces_file_id,
			source, source_ref_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NULL, $12, $13, $14, $15)
	`, f.ID, f.OwnerID, f.FolderID, f.Name, f.OriginalName, f.ObjectKey,
		f.MimeType, f.SizeBytes, f.Status, f.CreatedAt, f.UpdatedAt, f.UploadExpiresAt, f.ReplacesFileID,
		nullableSource(f.Source), f.SourceRefID)
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
			updated_at = $7, upload_expires_at = $8, source = $9, source_ref_id = $10
		WHERE id = $1 AND owner_id = $11 AND deleted_at IS NULL
	`, f.ID, f.Name, f.FolderID, f.MimeType, f.SizeBytes, f.Status, f.UpdatedAt, f.UploadExpiresAt,
		nullableSource(f.Source), f.SourceRefID, f.OwnerID)
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

func (s *Store) GetAliveFileByName(ctx context.Context, ownerID string, folderID *string, name string) (*file.File, error) {
	if folderID == nil {
		return scanFile(s.pool.QueryRow(ctx, fileSelect+` AND owner_id = $1 AND folder_id IS NULL AND name = $2`, ownerID, name))
	}
	return scanFile(s.pool.QueryRow(ctx, fileSelect+` AND owner_id = $1 AND folder_id = $2 AND name = $3`, ownerID, *folderID, name))
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

func (s *Store) CompleteUpload(ctx context.Context, ownerID, fileID string, size int64, contentType string, now time.Time) (*file.File, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var used, quota int64
	err = tx.QueryRow(ctx, `
		SELECT storage_used, storage_quota FROM users WHERE id = $1 FOR UPDATE
	`, ownerID).Scan(&used, &quota)
	if err != nil {
		return nil, err
	}
	if used+size > quota {
		return nil, apperr.QuotaExceeded
	}

	var f file.File
	err = tx.QueryRow(ctx, `
		UPDATE files
		SET status = 'READY',
			size_bytes = $2,
			mime_type = CASE WHEN $3 <> '' THEN $3 ELSE mime_type END,
			upload_expires_at = NULL,
			updated_at = $4
		WHERE id = $1 AND owner_id = $5 AND status = 'PENDING' AND deleted_at IS NULL
		RETURNING id, owner_id, folder_id, name, original_name, object_key,
			mime_type, size_bytes, status, created_at, updated_at, deleted_at, upload_expires_at, replaces_file_id,
			source, source_ref_id
	`, fileID, size, contentType, now, ownerID).Scan(
		&f.ID, &f.OwnerID, &f.FolderID, &f.Name, &f.OriginalName, &f.ObjectKey,
		&f.MimeType, &f.SizeBytes, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.DeletedAt, &f.UploadExpiresAt, &f.ReplacesFileID,
		&f.Source, &f.SourceRefID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.NotFound
	}
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperr.Conflict
		}
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE users SET storage_used = storage_used + $2, updated_at = $3 WHERE id = $1
	`, ownerID, size, now)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *Store) CompleteReplaceUpload(ctx context.Context, ownerID, pendingFileID, targetFileID string, size int64, contentType string, now time.Time) (*file.File, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var pending file.File
	err = tx.QueryRow(ctx, `
		SELECT id, object_key, mime_type FROM files
		WHERE id = $1 AND owner_id = $2 AND status = 'PENDING' AND deleted_at IS NULL
	`, pendingFileID, ownerID).Scan(&pending.ID, &pending.ObjectKey, &pending.MimeType)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.NotFound
	}
	if err != nil {
		return nil, err
	}

	var target file.File
	err = tx.QueryRow(ctx, `
		SELECT id, object_key, size_bytes, mime_type, name
		FROM files
		WHERE id = $1 AND owner_id = $2 AND status = 'READY' AND deleted_at IS NULL
		FOR UPDATE
	`, targetFileID, ownerID).Scan(&target.ID, &target.ObjectKey, &target.SizeBytes, &target.MimeType, &target.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.NotFound
	}
	if err != nil {
		return nil, err
	}

	sizeDelta := size - target.SizeBytes
	var used, quota int64
	err = tx.QueryRow(ctx, `
		SELECT storage_used, storage_quota FROM users WHERE id = $1 FOR UPDATE
	`, ownerID).Scan(&used, &quota)
	if err != nil {
		return nil, err
	}
	if sizeDelta > 0 && used+sizeDelta > quota {
		return nil, apperr.QuotaExceeded
	}

	versionID := auth.NewID()
	_, err = tx.Exec(ctx, `
		INSERT INTO file_versions (id, file_id, object_key, size_bytes, mime_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, versionID, target.ID, target.ObjectKey, target.SizeBytes, target.MimeType, now)
	if err != nil {
		return nil, err
	}

	// Delete the pending row first to free up pending.ObjectKey before updating target
	// to avoid duplicate key error on files_object_key_key unique constraint.
	_, err = tx.Exec(ctx, `DELETE FROM files WHERE id = $1 AND owner_id = $2`, pendingFileID, ownerID)
	if err != nil {
		return nil, err
	}

	var updated file.File
	err = tx.QueryRow(ctx, `
		UPDATE files
		SET object_key = $2,
			size_bytes = $3,
			mime_type = CASE WHEN $4 <> '' THEN $4 ELSE mime_type END,
			updated_at = $5
		WHERE id = $1 AND owner_id = $6
		RETURNING id, owner_id, folder_id, name, original_name, object_key,
			mime_type, size_bytes, status, created_at, updated_at, deleted_at, upload_expires_at, replaces_file_id,
			source, source_ref_id
	`, target.ID, pending.ObjectKey, size, contentType, now, ownerID).Scan(
		&updated.ID, &updated.OwnerID, &updated.FolderID, &updated.Name, &updated.OriginalName, &updated.ObjectKey,
		&updated.MimeType, &updated.SizeBytes, &updated.Status, &updated.CreatedAt, &updated.UpdatedAt, &updated.DeletedAt, &updated.UploadExpiresAt, &updated.ReplacesFileID,
		&updated.Source, &updated.SourceRefID,
	)
	if err != nil {
		return nil, err
	}

	if sizeDelta != 0 {
		_, err = tx.Exec(ctx, `
			UPDATE users SET storage_used = storage_used + $2, updated_at = $3 WHERE id = $1
		`, ownerID, sizeDelta, now)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &updated, nil
}

const fileSelect = `
	SELECT id, owner_id, folder_id, name, original_name, object_key,
		mime_type, size_bytes, status, created_at, updated_at, deleted_at, upload_expires_at, replaces_file_id,
		source, source_ref_id
	FROM files
	WHERE deleted_at IS NULL
`

// AddFavorite inserts a favorite row; re-favoriting keeps the original mark.
func (s *Store) AddFavorite(ctx context.Context, ownerID, fileID string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO favorites (id, owner_id, file_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (owner_id, file_id) DO NOTHING
	`, auth.NewID(), ownerID, fileID, at)
	return err
}

// RemoveFavorite deletes the favorite row if present; idempotent.
func (s *Store) RemoveFavorite(ctx context.Context, ownerID, fileID string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM favorites WHERE owner_id = $1 AND file_id = $2
	`, ownerID, fileID)
	return err
}

// ListFavorites returns the user's favorited files, newest favorite first.
func (s *Store) ListFavorites(ctx context.Context, ownerID string, limit int) ([]file.FavoriteFile, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT f.id, f.name, f.mime_type, f.size_bytes, f.updated_at, fav.created_at
		FROM favorites fav
		JOIN files f ON f.id = fav.file_id
		WHERE fav.owner_id = $1 AND f.deleted_at IS NULL AND f.status = 'READY' AND f.is_vault = FALSE
		ORDER BY fav.created_at DESC
		LIMIT $2
	`, ownerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []file.FavoriteFile
	for rows.Next() {
		var it file.FavoriteFile
		if err := rows.Scan(&it.ID, &it.Name, &it.MimeType, &it.SizeBytes, &it.UpdatedAt, &it.FavoritedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func scanFile(row pgx.Row) (*file.File, error) {
	var f file.File
	err := row.Scan(
		&f.ID, &f.OwnerID, &f.FolderID, &f.Name, &f.OriginalName, &f.ObjectKey,
		&f.MimeType, &f.SizeBytes, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.DeletedAt, &f.UploadExpiresAt, &f.ReplacesFileID,
		&f.Source, &f.SourceRefID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func nullableSource(source string) any {
	if source == "" {
		return "vault"
	}
	return source
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

func (r fileRepo) GetAliveByName(ctx context.Context, ownerID string, folderID *string, name string) (*file.File, error) {
	return r.store.GetAliveFileByName(ctx, ownerID, folderID, name)
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

func (r fileRepo) AddFavorite(ctx context.Context, ownerID, fileID string, at time.Time) error {
	return r.store.AddFavorite(ctx, ownerID, fileID, at)
}

func (r fileRepo) RemoveFavorite(ctx context.Context, ownerID, fileID string) error {
	return r.store.RemoveFavorite(ctx, ownerID, fileID)
}

func (r fileRepo) ListFavorites(ctx context.Context, ownerID string, limit int) ([]file.FavoriteFile, error) {
	return r.store.ListFavorites(ctx, ownerID, limit)
}

func (r fileRepo) CompleteUpload(ctx context.Context, ownerID, fileID string, size int64, contentType string, now time.Time) (*file.File, error) {
	return r.store.CompleteUpload(ctx, ownerID, fileID, size, contentType, now)
}

func (r fileRepo) CompleteReplaceUpload(ctx context.Context, ownerID, pendingFileID, targetFileID string, size int64, contentType string, now time.Time) (*file.File, error) {
	return r.store.CompleteReplaceUpload(ctx, ownerID, pendingFileID, targetFileID, size, contentType, now)
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
