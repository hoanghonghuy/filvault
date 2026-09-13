package postgres

import (
	"context"

	"filvault/internal/photo"
)

func (s *Store) ListTimelineFilesByType(ctx context.Context, ownerID string, before *photo.TimelineCursor, limit int, mediaType string) ([]photo.TimelineItem, error) {
	var beforeTime any
	var beforeID string
	if before != nil {
		beforeTime = before.CreatedAt
		beforeID = before.ID
	}

	rows, err := s.pool.Query(ctx, `
		SELECT f.id, f.name, f.mime_type, f.size_bytes, f.created_at, f.object_key,
			EXISTS (
				SELECT 1 FROM favorites fav
				WHERE fav.owner_id = $1 AND fav.file_id = f.id
			) AS is_favorite
		FROM files f
		WHERE f.owner_id = $1 AND f.deleted_at IS NULL AND f.status = 'READY' AND f.is_vault = FALSE
			AND f.mime_type = ANY($2::text[])
			AND f.mime_type LIKE ($3 || '/%')
			AND ($4::timestamptz IS NULL OR f.created_at < $4 OR (f.created_at = $4 AND f.id < $5))
		ORDER BY f.created_at DESC, f.id DESC
		LIMIT $6
	`, ownerID, photo.PhotoMIMESlice(), mediaType, beforeTime, beforeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []photo.TimelineItem
	for rows.Next() {
		var item photo.TimelineItem
		if err := rows.Scan(&item.ID, &item.Name, &item.MimeType, &item.SizeBytes, &item.CreatedAt, &item.ObjectKey, &item.IsFavorite); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r photoRepo) ListTimelineFilesByType(ctx context.Context, ownerID string, before *photo.TimelineCursor, limit int, mediaType string) ([]photo.TimelineItem, error) {
	return r.store.ListTimelineFilesByType(ctx, ownerID, before, limit, mediaType)
}
