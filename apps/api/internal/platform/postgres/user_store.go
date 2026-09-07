package postgres

import (
	"context"
	"errors"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateUser(ctx context.Context, u user.User) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (
			id, email, display_name, password_hash,
			email_verified_at, verification_code_hash, verification_expires_at,
			storage_used, storage_quota,
			image_thumbnails_enabled, video_thumbnails_enabled,
			trash_auto_delete_enabled, trash_retention_days,
			active_status_enabled,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, NULL, NULL,
			$6, $7,
			$8, $9,
			$10, $11,
			$12,
			$13, $14
		)
	`, u.ID, u.Email, u.DisplayName, u.PasswordHash,
		u.EmailVerifiedAt, u.StorageUsed, u.StorageQuota,
		u.ImageThumbnailsEnabled, u.VideoThumbnailsEnabled,
		u.TrashAutoDeleteEnabled, u.TrashRetentionDays,
		u.ActiveStatusEnabled,
		u.CreatedAt, u.UpdatedAt)
	if isUniqueViolation(err) {
		return apperr.Conflict
	}
	return err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.scanUser(s.pool.QueryRow(ctx, userSelect+` WHERE email = $1`, email))
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	return s.scanUser(s.pool.QueryRow(ctx, userSelect+` WHERE id = $1`, id))
}

func (s *Store) SetVerificationCode(ctx context.Context, userID, codeHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE users
		SET verification_code_hash = $2, verification_expires_at = $3, updated_at = now()
		WHERE id = $1
	`, userID, codeHash, expiresAt)
	return err
}

func (s *Store) MarkEmailVerified(ctx context.Context, userID string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE users
		SET email_verified_at = $2, verification_code_hash = NULL, verification_expires_at = NULL, updated_at = now()
		WHERE id = $1
	`, userID, at)
	return err
}

func (s *Store) UpdateTrashSettings(ctx context.Context, userID string, enabled bool, retentionDays int) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE users
		SET trash_auto_delete_enabled = $2, trash_retention_days = $3, updated_at = now()
		WHERE id = $1
	`, userID, enabled, retentionDays)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) UpdateThumbnailSettings(ctx context.Context, userID string, imageEnabled, videoEnabled bool) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE users
		SET image_thumbnails_enabled = $2, video_thumbnails_enabled = $3, updated_at = now()
		WHERE id = $1
	`, userID, imageEnabled, videoEnabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) UpdateDisplayName(ctx context.Context, userID, displayName string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE users SET display_name = $2, updated_at = now()
		WHERE id = $1
	`, userID, displayName)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE users SET password_hash = $2, updated_at = now()
		WHERE id = $1
	`, userID, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) UpdateActiveStatus(ctx context.Context, userID string, enabled bool) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE users SET active_status_enabled = $2, updated_at = now()
		WHERE id = $1
	`, userID, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) InsertRefreshToken(ctx context.Context, tok auth.RefreshToken) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tok.ID, tok.UserID, tok.TokenHash, tok.ExpiresAt, tok.RevokedAt, tok.CreatedAt)
	return err
}

func (s *Store) GetRefreshTokenByHash(ctx context.Context, hash string) (*auth.RefreshToken, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, hash)
	var tok auth.RefreshToken
	err := row.Scan(&tok.ID, &tok.UserID, &tok.TokenHash, &tok.ExpiresAt, &tok.RevokedAt, &tok.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tok, nil
}

func (s *Store) RevokeRefreshToken(ctx context.Context, hash string, at time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, hash, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.Unauthorized
	}
	return nil
}

func (s *Store) RevokeAllRefreshTokensForUser(ctx context.Context, userID string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, at)
	return err
}

const userSelect = `
	SELECT id, email, display_name, password_hash,
		email_verified_at, verification_code_hash, verification_expires_at,
		storage_used, storage_quota,
		image_thumbnails_enabled, video_thumbnails_enabled,
		trash_auto_delete_enabled, trash_retention_days,
		COALESCE(active_status_enabled, true),
		created_at, updated_at
	FROM users
`

func (s *Store) scanUser(row pgx.Row) (*user.User, error) {
	var u user.User
	var codeHash *string
	err := row.Scan(
		&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash,
		&u.EmailVerifiedAt, &codeHash, &u.VerificationExpiresAt,
		&u.StorageUsed, &u.StorageQuota,
		&u.ImageThumbnailsEnabled, &u.VideoThumbnailsEnabled,
		&u.TrashAutoDeleteEnabled, &u.TrashRetentionDays,
		&u.ActiveStatusEnabled,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if codeHash != nil {
		u.VerificationCodeHash = *codeHash
	}
	return &u, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
