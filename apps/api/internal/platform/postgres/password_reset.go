package postgres

import (
	"context"
	"errors"
	"time"

	"filvault/internal/apperr"

	"github.com/jackc/pgx/v5"
)

func (s *Store) SetPasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE users
		SET password_reset_token_hash = $2,
			password_reset_expires_at = $3,
			updated_at = now()
		WHERE id = $1
	`, userID, tokenHash, expiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound
	}
	return nil
}

func (s *Store) ConsumePasswordReset(ctx context.Context, tokenHash, passwordHash string, at time.Time) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET password_hash = $2,
			password_reset_token_hash = NULL,
			password_reset_expires_at = NULL,
			updated_at = $3
		WHERE password_reset_token_hash = $1
			AND password_reset_expires_at > $3
		RETURNING id
	`, tokenHash, passwordHash, at).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", apperr.Validation
	}
	if err != nil {
		return "", err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, at); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return userID, nil
}
