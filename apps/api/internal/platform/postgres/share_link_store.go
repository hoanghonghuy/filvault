package postgres

import (
	"context"
	"errors"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/sharelink"

	"github.com/jackc/pgx/v5"
)

const shareLinkSelect = `
	SELECT l.id, l.owner_id, l.file_id, l.token_hash, l.expires_at, l.revoked_at, l.created_at,
	       f.name, f.mime_type, f.size_bytes, f.object_key, f.status
	FROM share_links l
	JOIN files f ON f.id = l.file_id
`

func scanShareLink(row pgx.Row) (*sharelink.ShareLink, error) {
	var l sharelink.ShareLink
	err := row.Scan(
		&l.ID, &l.OwnerID, &l.FileID, &l.TokenHash, &l.ExpiresAt, &l.RevokedAt, &l.CreatedAt,
		&l.FileInfo.Name, &l.FileInfo.MimeType, &l.FileInfo.SizeBytes, &l.FileInfo.ObjectKey, &l.FileInfo.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

type shareLinkRepo struct {
	store *Store
}

func NewShareLinkRepository(store *Store) sharelink.Repository {
	return shareLinkRepo{store: store}
}

func (r shareLinkRepo) Create(ctx context.Context, link sharelink.ShareLink) error {
	if link.ID == "" {
		link.ID = auth.NewID()
	}
	_, err := r.store.pool.Exec(ctx, `
		INSERT INTO share_links (id, owner_id, file_id, token_hash, expires_at, revoked_at, created_at, recipient_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULL)
	`, link.ID, link.OwnerID, link.FileID, link.TokenHash, link.ExpiresAt, link.RevokedAt, link.CreatedAt)
	if isUniqueViolation(err) {
		return apperr.Conflict
	}
	return err
}

func (r shareLinkRepo) CountActiveByOwner(ctx context.Context, ownerID string) (int, error) {
	var n int
	err := r.store.pool.QueryRow(ctx, `
		SELECT count(*) FROM share_links
		WHERE owner_id = $1 AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > now())
	`, ownerID).Scan(&n)
	return n, err
}

func (r shareLinkRepo) FileIsShareable(ctx context.Context, ownerID, fileID string) (bool, error) {
	var ok bool
	err := r.store.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM files
			WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL AND status = 'READY'
		)
	`, fileID, ownerID).Scan(&ok)
	return ok, err
}

func (r shareLinkRepo) RevokeActiveForFile(ctx context.Context, fileID string, at time.Time) error {
	_, err := r.store.pool.Exec(ctx, `
		UPDATE share_links SET revoked_at = $2
		WHERE file_id = $1 AND revoked_at IS NULL
	`, fileID, at)
	return err
}

func (r shareLinkRepo) ListActiveByOwner(ctx context.Context, ownerID string) ([]sharelink.ShareLink, error) {
	rows, err := r.store.pool.Query(ctx, shareLinkSelect+`
		WHERE l.owner_id = $1 AND l.revoked_at IS NULL AND (l.expires_at IS NULL OR l.expires_at > now())
		  AND f.deleted_at IS NULL AND f.status = 'READY'
		ORDER BY l.created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []sharelink.ShareLink
	for rows.Next() {
		l, err := scanShareLink(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (r shareLinkRepo) FindActiveByTokenHash(ctx context.Context, tokenHash string) (*sharelink.ShareLink, error) {
	return scanShareLink(r.store.pool.QueryRow(ctx, shareLinkSelect+`
		WHERE l.token_hash = $1 AND l.revoked_at IS NULL AND (l.expires_at IS NULL OR l.expires_at > now())
		  AND f.deleted_at IS NULL AND f.status = 'READY'
	`, tokenHash))
}

func (r shareLinkRepo) FileName(ctx context.Context, ownerID, fileID string) (string, error) {
	var name string
	err := r.store.pool.QueryRow(ctx, `
		SELECT name FROM files WHERE id = $1 AND owner_id = $2
	`, fileID, ownerID).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return name, nil
}
