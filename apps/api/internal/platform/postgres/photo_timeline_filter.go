package postgres

import (
	"context"
	"time"

	"filvault/internal/photo"
)

func (s *Store) ListTimelineFilesByType(ctx context.Context, ownerID string, before *time.Time, limit int, mediaType string) ([]photo.TimelineItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, mime_type, size_bytes, created_at, object_key
		FROM files
		WHERE owner_id = $1 AND deleted_at IS NULL AND status = 'READY' AND is_vault = FALSE
			AND mime_type = ANY($2::text[])
			AND mime_type LIKE ($3 || '/%')
			AND ($4::timestamptz IS NULL OR created_at < $4)
		ORDER BY created_at DESC
		LIMIT $5
	`, ownerID, photo.PhotoMIMESlice(), mediaType, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTimelineItems(rows)
}

func (r photoRepo) ListTimelineFilesByType(ctx context.Context, ownerID string, before *time.Time, limit int, mediaType string) ([]photo.TimelineItem, error) {
	return r.store.ListTimelineFilesByType(ctx, ownerID, before, limit, mediaType)
}
