package postgres

import (
	"context"
	"errors"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/file"
	"filvault/internal/folder"
	"filvault/internal/trash"
	"filvault/internal/user"

	"github.com/jackc/pgx/v5"
)

const fileSelectAny = `
	SELECT id, owner_id, folder_id, name, original_name, object_key,
		mime_type, size_bytes, status, created_at, updated_at, deleted_at, upload_expires_at, replaces_file_id,
		source, source_ref_id
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

func (s *Store) PurgeFile(ctx context.Context, ownerID, id string) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	attachments, err := fileAttachmentEvents(ctx, tx, id)
	if err != nil {
		return "", err
	}
	var objectKey string
	var sizeBytes int64
	var status string
	err = tx.QueryRow(ctx, `
		SELECT object_key, size_bytes, status
		FROM files
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NOT NULL
		FOR UPDATE
	`, id, ownerID).Scan(&objectKey, &sizeBytes, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.QueryRow(ctx, `
			SELECT object_key
			FROM file_purge_jobs
			WHERE owner_id = $1 AND file_id = $2 AND completed_at IS NULL
		`, ownerID, id).Scan(&objectKey); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "", apperr.NotFound
			}
			return "", err
		}
		if err := tx.Commit(ctx); err != nil {
			return "", err
		}
		return objectKey, nil
	}
	if err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM files WHERE id = $1 AND owner_id = $2`, id, ownerID); err != nil {
		return "", err
	}
	for _, attachment := range attachments {
		if err := insertChatEvent(ctx, tx, attachment.ConversationID, "attachment.updated", attachment.AttachmentID, time.Now().UTC()); err != nil {
			return "", err
		}
	}
	if status == file.StatusReady && sizeBytes > 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE users
			SET storage_used = storage_used - $2, updated_at = now()
			WHERE id = $1
		`, ownerID, sizeBytes); err != nil {
			return "", err
		}
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO file_purge_jobs (id, owner_id, file_id, object_key, created_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (owner_id, file_id) WHERE completed_at IS NULL DO NOTHING
	`, auth.NewID(), ownerID, id, objectKey); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return objectKey, nil
}

func (s *Store) CompleteFilePurge(ctx context.Context, ownerID, fileID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE file_purge_jobs
		SET completed_at = now()
		WHERE owner_id = $1 AND file_id = $2 AND completed_at IS NULL
	`, ownerID, fileID)
	if err == nil && tag.RowsAffected() == 0 {
		var exists bool
		if lookupErr := s.pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM file_purge_jobs
				WHERE owner_id = $1 AND file_id = $2 AND completed_at IS NOT NULL
			)
		`, ownerID, fileID).Scan(&exists); lookupErr != nil {
			return lookupErr
		} else if !exists {
			return apperr.NotFound
		}
	}
	return err
}

func (s *Store) ClaimPurgeJob(ctx context.Context, workerID string, lease time.Duration) (*trash.PurgeJob, error) {
	var job trash.PurgeJob
	leaseMillis := lease.Milliseconds()
	if leaseMillis <= 0 {
		leaseMillis = time.Minute.Milliseconds()
	}
	err := s.pool.QueryRow(ctx, `
		UPDATE file_purge_jobs
		SET claimed_at = now(), claimed_by = $1, attempts = attempts + 1
		WHERE id = (
			SELECT id FROM file_purge_jobs
			WHERE completed_at IS NULL AND failed_at IS NULL
				AND next_attempt_at <= now()
				AND (claimed_at IS NULL OR claimed_at < now() - ($2::bigint * INTERVAL '1 millisecond'))
			ORDER BY next_attempt_at ASC, created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, owner_id, file_id, object_key, attempts
		`, workerID, leaseMillis).Scan(&job.ID, &job.OwnerID, &job.FileID, &job.ObjectKey, &job.Attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *Store) CompletePurgeJob(ctx context.Context, jobID, workerID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE file_purge_jobs
		SET completed_at = now(), claimed_at = NULL, claimed_by = NULL
		WHERE id = $1 AND claimed_by = $2 AND completed_at IS NULL
	`, jobID, workerID)
	if err == nil && tag.RowsAffected() == 0 {
		return apperr.InvalidState
	}
	return err
}

func (s *Store) FailPurgeJob(ctx context.Context, jobID, workerID, lastError string, retryAt time.Time, dead bool) error {
	var deadAt any
	if dead {
		deadAt = time.Now().UTC()
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE file_purge_jobs
		SET next_attempt_at = $4, last_error = $3, failed_at = $5,
			claimed_at = NULL, claimed_by = NULL
		WHERE id = $1 AND claimed_by = $2 AND completed_at IS NULL
	`, jobID, workerID, lastError, retryAt, deadAt)
	if err == nil && tag.RowsAffected() == 0 {
		return apperr.InvalidState
	}
	return err
}

func (s *Store) ClaimUploadCleanupJob(ctx context.Context, workerID string, lease time.Duration) (*trash.UploadCleanupJob, error) {
	var job trash.UploadCleanupJob
	leaseMillis := lease.Milliseconds()
	if leaseMillis <= 0 {
		leaseMillis = time.Minute.Milliseconds()
	}
	err := s.pool.QueryRow(ctx, `
		UPDATE file_upload_cleanup_jobs
		SET claimed_at = now(), claimed_by = $1, attempts = attempts + 1
		WHERE id = (
			SELECT id FROM file_upload_cleanup_jobs
			WHERE completed_at IS NULL AND failed_at IS NULL
				AND next_attempt_at <= now()
				AND (claimed_at IS NULL OR claimed_at < now() - ($2::bigint * INTERVAL '1 millisecond'))
			ORDER BY next_attempt_at ASC, created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, owner_id, file_id, object_key, attempts
	`, workerID, leaseMillis).Scan(&job.ID, &job.OwnerID, &job.FileID, &job.ObjectKey, &job.Attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *Store) CompleteUploadCleanupJob(ctx context.Context, jobID, workerID, ownerID, fileID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `
		UPDATE file_upload_cleanup_jobs
		SET completed_at = now(), claimed_at = NULL, claimed_by = NULL
		WHERE id = $1 AND claimed_by = $2 AND completed_at IS NULL
	`, jobID, workerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.InvalidState
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM files
		WHERE id = $1 AND owner_id = $2 AND status = 'PENDING'
	`, fileID, ownerID); err != nil {
		return err
	}
	err = tx.Commit(ctx)
	return err
}

func (s *Store) FailUploadCleanupJob(ctx context.Context, jobID, workerID, lastError string, retryAt time.Time, dead bool) error {
	var deadAt any
	if dead {
		deadAt = time.Now().UTC()
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE file_upload_cleanup_jobs
		SET next_attempt_at = $4, last_error = $3, failed_at = $5,
			claimed_at = NULL, claimed_by = NULL
		WHERE id = $1 AND claimed_by = $2 AND completed_at IS NULL
	`, jobID, workerID, lastError, retryAt, deadAt)
	if err == nil && tag.RowsAffected() == 0 {
		return apperr.InvalidState
	}
	return err
}

func (s *Store) QueueExpiredUploadCleanupJobs(ctx context.Context, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO file_upload_cleanup_jobs (id, owner_id, file_id, object_key, created_at)
		SELECT f.id, f.owner_id, f.id, f.object_key, COALESCE(f.upload_expires_at, f.created_at)
		FROM files f
		WHERE f.status = 'PENDING'
			AND f.upload_expires_at IS NOT NULL
			AND f.upload_expires_at <= $1
		ON CONFLICT (file_id) DO NOTHING
	`, now)
	return err
}

func (s *Store) SoftDeleteFile(ctx context.Context, ownerID, id string, at time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	attachments, err := fileAttachmentEvents(ctx, tx, id)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
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
	for _, attachment := range attachments {
		if err := insertChatEvent(ctx, tx, attachment.ConversationID, "attachment.updated", attachment.AttachmentID, at); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) RestoreFile(ctx context.Context, ownerID, id string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	attachments, err := fileAttachmentEvents(ctx, tx, id)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE files SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NOT NULL
	`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	for _, attachment := range attachments {
		if err := insertChatEvent(ctx, tx, attachment.ConversationID, "attachment.updated", attachment.AttachmentID, time.Now().UTC()); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type fileAttachmentEvent struct {
	ConversationID string
	AttachmentID   string
}

func fileAttachmentEvents(ctx context.Context, tx pgx.Tx, fileID string) ([]fileAttachmentEvent, error) {
	rows, err := tx.Query(ctx, `
		SELECT m.conversation_id, a.id
		FROM message_attachments a
		JOIN messages m ON m.id = a.message_id
		WHERE a.file_id = $1
	`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []fileAttachmentEvent
	for rows.Next() {
		var event fileAttachmentEvent
		if err := rows.Scan(&event.ConversationID, &event.AttachmentID); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
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
			&f.Source, &f.SourceRefID,
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

func (r trashRepo) PurgeFile(ctx context.Context, ownerID, id string) (string, error) {
	return r.store.PurgeFile(ctx, ownerID, id)
}

func (r trashRepo) CompleteFilePurge(ctx context.Context, ownerID, fileID string) error {
	return r.store.CompleteFilePurge(ctx, ownerID, fileID)
}

func (r trashRepo) ClaimPurgeJob(ctx context.Context, workerID string, lease time.Duration) (*trash.PurgeJob, error) {
	return r.store.ClaimPurgeJob(ctx, workerID, lease)
}

func (r trashRepo) CompletePurgeJob(ctx context.Context, jobID, workerID string) error {
	return r.store.CompletePurgeJob(ctx, jobID, workerID)
}

func (r trashRepo) FailPurgeJob(ctx context.Context, jobID, workerID, lastError string, retryAt time.Time, dead bool) error {
	return r.store.FailPurgeJob(ctx, jobID, workerID, lastError, retryAt, dead)
}

func (r trashRepo) ClaimUploadCleanupJob(ctx context.Context, workerID string, lease time.Duration) (*trash.UploadCleanupJob, error) {
	return r.store.ClaimUploadCleanupJob(ctx, workerID, lease)
}

func (r trashRepo) CompleteUploadCleanupJob(ctx context.Context, jobID, workerID, ownerID, fileID string) error {
	return r.store.CompleteUploadCleanupJob(ctx, jobID, workerID, ownerID, fileID)
}

func (r trashRepo) FailUploadCleanupJob(ctx context.Context, jobID, workerID, lastError string, retryAt time.Time, dead bool) error {
	return r.store.FailUploadCleanupJob(ctx, jobID, workerID, lastError, retryAt, dead)
}

func (r trashRepo) QueueExpiredUploadCleanupJobs(ctx context.Context, now time.Time) error {
	return r.store.QueueExpiredUploadCleanupJobs(ctx, now)
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
