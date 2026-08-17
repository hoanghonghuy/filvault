package postgres

import (
	"context"

	"filnest/internal/search"
)

func (s *Store) SearchFolders(ctx context.Context, ownerID, pattern string, limit int) ([]search.FolderHit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, parent_id, name, created_at, updated_at
		FROM folders
		WHERE owner_id = $1 AND deleted_at IS NULL
			AND name ILIKE '%' || $2 || '%' ESCAPE '\'
		ORDER BY name ASC
		LIMIT $3
	`, ownerID, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []search.FolderHit
	for rows.Next() {
		var f search.FolderHit
		if err := rows.Scan(&f.ID, &f.ParentID, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Store) SearchFiles(ctx context.Context, ownerID, pattern string, limit int) ([]search.FileHit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, mime_type, size_bytes, updated_at
		FROM files
		WHERE owner_id = $1 AND deleted_at IS NULL AND status = 'READY'
			AND name ILIKE '%' || $2 || '%' ESCAPE '\'
		ORDER BY name ASC
		LIMIT $3
	`, ownerID, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []search.FileHit
	for rows.Next() {
		var f search.FileHit
		if err := rows.Scan(&f.ID, &f.Name, &f.MimeType, &f.SizeBytes, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

type searchRepo struct {
	store *Store
}

func NewSearchRepository(store *Store) search.Repository {
	return searchRepo{store: store}
}

func (r searchRepo) SearchFolders(ctx context.Context, ownerID, pattern string, limit int) ([]search.FolderHit, error) {
	return r.store.SearchFolders(ctx, ownerID, pattern, limit)
}

func (r searchRepo) SearchFiles(ctx context.Context, ownerID, pattern string, limit int) ([]search.FileHit, error) {
	return r.store.SearchFiles(ctx, ownerID, pattern, limit)
}

var _ search.Repository = searchRepo{}
