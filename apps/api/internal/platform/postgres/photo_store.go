package postgres

import (
	"context"
	"errors"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/photo"

	"github.com/jackc/pgx/v5"
)

func (s *Store) ListTimelineFiles(ctx context.Context, ownerID string, before *time.Time, limit int) ([]photo.TimelineItem, error) {
	mimes := photo.PhotoMIMESlice()
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, mime_type, size_bytes, created_at
		FROM files
		WHERE owner_id = $1 AND deleted_at IS NULL AND status = 'READY'
			AND mime_type = ANY($2::text[])
			AND ($3::timestamptz IS NULL OR created_at < $3)
		ORDER BY created_at DESC
		LIMIT $4
	`, ownerID, mimes, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTimelineItems(rows)
}

func (s *Store) GetPhotoFile(ctx context.Context, ownerID, fileID string) (*photo.TimelineItem, error) {
	var item photo.TimelineItem
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, mime_type, size_bytes, created_at
		FROM files
		WHERE owner_id = $1 AND id = $2 AND deleted_at IS NULL AND status = 'READY'
	`, ownerID, fileID).Scan(&item.ID, &item.Name, &item.MimeType, &item.SizeBytes, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.NotFound
	}
	if err != nil {
		return nil, err
	}
	if !photo.IsPhotoMime(item.MimeType) {
		return nil, apperr.Validation
	}
	return &item, nil
}

func (s *Store) CreateAlbum(ctx context.Context, a photo.Album) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO albums (id, owner_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, a.ID, a.OwnerID, a.Name, a.CreatedAt, a.UpdatedAt)
	if isUniqueViolation(err) {
		return apperr.Conflict
	}
	return err
}

func (s *Store) GetAlbum(ctx context.Context, ownerID, id string) (*photo.Album, error) {
	var a photo.Album
	err := s.pool.QueryRow(ctx, `
		SELECT a.id, a.owner_id, a.name, a.created_at, a.updated_at,
			COUNT(ai.file_id) AS item_count
		FROM albums a
		LEFT JOIN album_items ai ON ai.album_id = a.id
		WHERE a.owner_id = $1 AND a.id = $2
		GROUP BY a.id
	`, ownerID, id).Scan(&a.ID, &a.OwnerID, &a.Name, &a.CreatedAt, &a.UpdatedAt, &a.ItemCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) ListAlbums(ctx context.Context, ownerID string) ([]photo.Album, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.owner_id, a.name, a.created_at, a.updated_at,
			COUNT(ai.file_id) AS item_count
		FROM albums a
		LEFT JOIN album_items ai ON ai.album_id = a.id
		WHERE a.owner_id = $1
		GROUP BY a.id
		ORDER BY a.name ASC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []photo.Album
	for rows.Next() {
		var a photo.Album
		if err := rows.Scan(&a.ID, &a.OwnerID, &a.Name, &a.CreatedAt, &a.UpdatedAt, &a.ItemCount); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) UpdateAlbum(ctx context.Context, a photo.Album) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE albums SET name = $2, updated_at = $3
		WHERE id = $1 AND owner_id = $4
	`, a.ID, a.Name, a.UpdatedAt, a.OwnerID)
	if isUniqueViolation(err) {
		return apperr.Conflict
	}
	return err
}

func (s *Store) DeleteAlbum(ctx context.Context, ownerID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM albums WHERE id = $1 AND owner_id = $2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) AddAlbumItem(ctx context.Context, albumID, fileID string, position int, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO album_items (album_id, file_id, position, added_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (album_id, file_id) DO NOTHING
	`, albumID, fileID, position, at)
	return err
}

func (s *Store) RemoveAlbumItem(ctx context.Context, albumID, fileID string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM album_items WHERE album_id = $1 AND file_id = $2
	`, albumID, fileID)
	return err
}

func (s *Store) ListAlbumItems(ctx context.Context, ownerID, albumID string) ([]photo.TimelineItem, error) {
	mimes := photo.PhotoMIMESlice()
	rows, err := s.pool.Query(ctx, `
		SELECT f.id, f.name, f.mime_type, f.size_bytes, f.created_at
		FROM album_items ai
		JOIN albums a ON a.id = ai.album_id
		JOIN files f ON f.id = ai.file_id
		WHERE a.owner_id = $1 AND ai.album_id = $2
			AND f.deleted_at IS NULL AND f.status = 'READY'
			AND f.mime_type = ANY($3::text[])
		ORDER BY ai.position ASC, ai.added_at ASC
	`, ownerID, albumID, mimes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTimelineItems(rows)
}

func scanTimelineItems(rows pgx.Rows) ([]photo.TimelineItem, error) {
	var out []photo.TimelineItem
	for rows.Next() {
		var item photo.TimelineItem
		if err := rows.Scan(&item.ID, &item.Name, &item.MimeType, &item.SizeBytes, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

type photoRepo struct {
	store *Store
}

func NewPhotoRepository(store *Store) photo.Repository {
	return photoRepo{store: store}
}

func (r photoRepo) ListTimelineFiles(ctx context.Context, ownerID string, before *time.Time, limit int) ([]photo.TimelineItem, error) {
	return r.store.ListTimelineFiles(ctx, ownerID, before, limit)
}

func (r photoRepo) GetPhotoFile(ctx context.Context, ownerID, fileID string) (*photo.TimelineItem, error) {
	return r.store.GetPhotoFile(ctx, ownerID, fileID)
}

func (r photoRepo) CreateAlbum(ctx context.Context, a photo.Album) error {
	return r.store.CreateAlbum(ctx, a)
}

func (r photoRepo) GetAlbum(ctx context.Context, ownerID, id string) (*photo.Album, error) {
	return r.store.GetAlbum(ctx, ownerID, id)
}

func (r photoRepo) ListAlbums(ctx context.Context, ownerID string) ([]photo.Album, error) {
	return r.store.ListAlbums(ctx, ownerID)
}

func (r photoRepo) UpdateAlbum(ctx context.Context, a photo.Album) error {
	return r.store.UpdateAlbum(ctx, a)
}

func (r photoRepo) DeleteAlbum(ctx context.Context, ownerID, id string) error {
	return r.store.DeleteAlbum(ctx, ownerID, id)
}

func (r photoRepo) AddAlbumItem(ctx context.Context, albumID, fileID string, position int, at time.Time) error {
	return r.store.AddAlbumItem(ctx, albumID, fileID, position, at)
}

func (r photoRepo) RemoveAlbumItem(ctx context.Context, albumID, fileID string) error {
	return r.store.RemoveAlbumItem(ctx, albumID, fileID)
}

func (r photoRepo) ListAlbumItems(ctx context.Context, ownerID, albumID string) ([]photo.TimelineItem, error) {
	return r.store.ListAlbumItems(ctx, ownerID, albumID)
}

var _ photo.Repository = photoRepo{}
